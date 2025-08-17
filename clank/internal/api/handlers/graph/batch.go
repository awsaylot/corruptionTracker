package graph

import (
	"clank/internal/db"
	"clank/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// BatchCreateNodes creates multiple nodes in a single transaction
func BatchCreateNodes(c *gin.Context) {
	var nodes []models.Node
	if err := c.ShouldBindJSON(&nodes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		// Build UNWIND query
		query := `
			UNWIND $nodes as node
			CREATE (n:%s)
			SET n = node.props
			RETURN collect(n) as nodes
		`

		// Group nodes by type
		nodesByType := make(map[string][]map[string]interface{})
		for _, node := range nodes {
			// Add timestamps
			now := time.Now()
			if node.Props == nil {
				node.Props = make(map[string]any)
			}
			node.Props["created_at"] = now
			node.Props["updated_at"] = now

			nodeData := map[string]interface{}{
				"props": node.Props,
			}
			nodesByType[node.Type] = append(nodesByType[node.Type], nodeData)
		}

		var createdNodes []models.Node
		// Execute query for each node type
		for nodeType, nodeList := range nodesByType {
			params := map[string]interface{}{
				"nodes": nodeList,
			}

			result, err := tx.Run(fmt.Sprintf(query, nodeType), params)
			if err != nil {
				return nil, err
			}

			if !result.Next() {
				return nil, fmt.Errorf("no nodes created for type %s", nodeType)
			}

			record := result.Record()
			nodesCreated := record.Values[0].([]interface{})

			for _, n := range nodesCreated {
				node := n.(neo4j.Node)
				createdNodes = append(createdNodes, models.Node{
					ID:    fmt.Sprint(node.Id),
					Type:  node.Labels[0],
					Props: node.Props,
				})
			}
		}

		return createdNodes, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// BatchDeleteNodes deletes multiple nodes in a single transaction
func BatchDeleteNodes(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			UNWIND $ids as id
			MATCH (n)
			WHERE ID(n) = id
			DETACH DELETE n
		`
		params := map[string]interface{}{
			"ids": ids,
		}

		result, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		summary, err := result.Consume()
		if err != nil {
			return nil, err
		}

		if summary.Counters().NodesDeleted() == 0 {
			return nil, fmt.Errorf("no nodes were deleted")
		}

		return map[string]interface{}{
			"deleted": summary.Counters().NodesDeleted(),
		}, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
