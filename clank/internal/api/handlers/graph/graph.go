package graph

import (
	"clank/internal/db"
	"clank/internal/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// GetShortestPath finds the shortest path between two nodes
func GetShortestPath(c *gin.Context) {
	fromId := c.Query("from")
	toId := c.Query("to")
	if fromId == "" || toId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both from and to node IDs are required"})
		return
	}

	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH path = shortestPath((a)-[*]-(b))
			WHERE ID(a) = $fromId AND ID(b) = $toId
			RETURN path
		`
		params := map[string]interface{}{
			"fromId": fromId,
			"toId":   toId,
		}

		result, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("no path found")
		}

		record := result.Record()
		path := record.Values[0].(neo4j.Path)

		// Convert path to a more readable format
		nodes := make([]models.Node, len(path.Nodes))
		for i, node := range path.Nodes {
			nodes[i] = models.Node{
				ID:    fmt.Sprint(node.Id),
				Type:  node.Labels[0],
				Props: node.Props,
			}
		}

		relationships := make([]models.Relationship, len(path.Relationships))
		for i, rel := range path.Relationships {
			relationships[i] = models.Relationship{
				ID:     fmt.Sprint(rel.Id),
				FromID: fmt.Sprint(rel.StartId),
				ToID:   fmt.Sprint(rel.EndId),
				Type:   rel.Type,
				Props:  rel.Props,
			}
		}

		return map[string]interface{}{
			"nodes":         nodes,
			"relationships": relationships,
			"length":        len(relationships),
		}, nil
	})

	if err != nil {
		if err.Error() == "no path found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSubgraph returns a subgraph centered around a node with a specified depth
func GetSubgraph(c *gin.Context) {
	nodeId := c.Param("nodeId")
	depth := c.DefaultQuery("depth", "2")
	d, err := strconv.Atoi(depth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid depth parameter"})
		return
	}

	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH path = (n)-[*0..%d]-(m)
			WHERE ID(n) = $nodeId
			RETURN n as center, collect(DISTINCT path) as paths
		`
		params := map[string]interface{}{
			"nodeId": nodeId,
		}

		result, err := tx.Run(fmt.Sprintf(query, d), params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("node not found")
		}

		record := result.Record()
		centerNode := record.Values[0].(neo4j.Node)
		paths := record.Values[1].([]interface{})

		// Track unique nodes and relationships
		nodesMap := make(map[int64]models.Node)
		relsMap := make(map[int64]models.Relationship)

		// Add center node
		nodesMap[centerNode.Id] = models.Node{
			ID:    fmt.Sprint(centerNode.Id),
			Type:  centerNode.Labels[0],
			Props: centerNode.Props,
		}

		// Process all paths
		for _, p := range paths {
			path := p.(neo4j.Path)

			// Add nodes
			for _, node := range path.Nodes {
				if _, exists := nodesMap[node.Id]; !exists {
					nodesMap[node.Id] = models.Node{
						ID:    fmt.Sprint(node.Id),
						Type:  node.Labels[0],
						Props: node.Props,
					}
				}
			}

			// Add relationships
			for _, rel := range path.Relationships {
				if _, exists := relsMap[rel.Id]; !exists {
					relsMap[rel.Id] = models.Relationship{
						ID:     fmt.Sprint(rel.Id),
						FromID: fmt.Sprint(rel.StartId),
						ToID:   fmt.Sprint(rel.EndId),
						Type:   rel.Type,
						Props:  rel.Props,
					}
				}
			}
		}

		// Convert maps to slices
		nodes := make([]models.Node, 0, len(nodesMap))
		for _, node := range nodesMap {
			nodes = append(nodes, node)
		}

		relationships := make([]models.Relationship, 0, len(relsMap))
		for _, rel := range relsMap {
			relationships = append(relationships, rel)
		}

		return map[string]interface{}{
			"center":        nodesMap[centerNode.Id],
			"nodes":         nodes,
			"relationships": relationships,
			"depth":         d,
			"nodeCount":     len(nodes),
			"relCount":      len(relationships),
		}, nil
	})

	if err != nil {
		if err.Error() == "node not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
