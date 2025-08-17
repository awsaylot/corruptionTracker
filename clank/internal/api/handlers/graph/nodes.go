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

// GetAllNodes returns all nodes in the database
func GetAllNodes(c *gin.Context) {
	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)
			RETURN n
		`
		result, err := tx.Run(query, nil)
		if err != nil {
			return nil, err
		}

		var nodes []models.Node
		for result.Next() {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)

			nodes = append(nodes, models.Node{
				ID:    fmt.Sprint(node.Id),
				Type:  node.Labels[0],
				Props: node.Props,
			})
		}

		return nodes, nil
	})

	if err != nil {
		handleDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateNode creates a new node in the database
func CreateNode(c *gin.Context) {
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add timestamps
	now := time.Now()
	if node.Props == nil {
		node.Props = make(map[string]any)
	}
	node.Props["created_at"] = now
	node.Props["updated_at"] = now

	result, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			CREATE (n:%s $props)
			RETURN n
		`
		params := map[string]interface{}{
			"props": node.Props,
		}

		result, err := tx.Run(fmt.Sprintf(query, node.Type), params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("no node created")
		}

		record := result.Record()
		createdNode := record.Values[0].(neo4j.Node)

		return models.Node{
			ID:    fmt.Sprint(createdNode.Id),
			Type:  createdNode.Labels[0],
			Props: createdNode.Props,
		}, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetNode returns a specific node by ID
func GetNode(c *gin.Context) {
	id := c.Param("id")

	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)
			WHERE ID(n) = $id
			RETURN n
		`
		params := map[string]interface{}{
			"id": id,
		}

		result, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("node not found")
		}

		record := result.Record()
		node := record.Values[0].(neo4j.Node)

		return models.Node{
			ID:    fmt.Sprint(node.Id),
			Type:  node.Labels[0],
			Props: node.Props,
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

// UpdateNode updates a node by ID
func UpdateNode(c *gin.Context) {
	id := c.Param("id")
	var update models.Node
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add updated timestamp
	if update.Props == nil {
		update.Props = make(map[string]any)
	}
	update.Props["updated_at"] = time.Now()

	result, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)
			WHERE ID(n) = $id
			SET n += $props
			RETURN n
		`
		params := map[string]interface{}{
			"id":    id,
			"props": update.Props,
		}

		result, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("node not found")
		}

		record := result.Record()
		node := record.Values[0].(neo4j.Node)

		return models.Node{
			ID:    fmt.Sprint(node.Id),
			Type:  node.Labels[0],
			Props: node.Props,
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

// DeleteNode deletes a node by ID
func DeleteNode(c *gin.Context) {
	id := c.Param("id")

	_, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)
			WHERE ID(n) = $id
			DELETE n
		`
		params := map[string]interface{}{
			"id": id,
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
			return nil, fmt.Errorf("node not found")
		}

		return nil, nil
	})

	if err != nil {
		if err.Error() == "node not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchNodes searches nodes based on properties
func SearchNodes(c *gin.Context) {
	query := c.Query("q")
	nodeType := c.Query("type")

	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		var cypher string
		params := map[string]interface{}{
			"query": "(?i).*" + query + ".*", // Case-insensitive regex
		}

		if nodeType != "" {
			cypher = `
				MATCH (n:%s)
				WHERE any(prop in keys(n) WHERE n[prop] =~ $query)
				RETURN n
			`
			cypher = fmt.Sprintf(cypher, nodeType)
		} else {
			cypher = `
				MATCH (n)
				WHERE any(prop in keys(n) WHERE n[prop] =~ $query)
				RETURN n
			`
		}

		result, err := tx.Run(cypher, params)
		if err != nil {
			return nil, err
		}

		var nodes []models.Node
		for result.Next() {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)

			nodes = append(nodes, models.Node{
				ID:    fmt.Sprint(node.Id),
				Type:  node.Labels[0],
				Props: node.Props,
			})
		}

		return nodes, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetNetwork returns the entire graph network
func GetNetwork(c *gin.Context) {
	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)
			OPTIONAL MATCH (n)-[r]-(m)
			RETURN n, collect({node: m, relationship: r}) as connections
		`

		result, err := tx.Run(query, nil)
		if err != nil {
			return nil, err
		}

		var network []models.NodeWithConnections
		for result.Next() {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)
			connections := record.Values[1].([]interface{})

			nodeWithConn := models.NodeWithConnections{
				ID:         fmt.Sprint(node.Id),
				Type:       node.Labels[0],
				Properties: node.Props,
			}

			for _, conn := range connections {
				if conn == nil {
					continue
				}

				connMap := conn.(map[string]interface{})

				// Skip if either node or relationship is nil
				if connMap["node"] == nil || connMap["relationship"] == nil {
					continue
				}

				connNode, ok := connMap["node"].(neo4j.Node)
				if !ok {
					continue
				}

				rel, ok := connMap["relationship"].(neo4j.Relationship)
				if !ok {
					continue
				}

				connection := models.Connection{
					ID:         fmt.Sprint(connNode.Id),
					Type:       connNode.Labels[0],
					Properties: connNode.Props,
				}
				connection.Relationship.Type = rel.Type
				connection.Relationship.Properties = rel.Props
				connection.Relationship.Direction = "outgoing"
				if rel.StartId == node.Id {
					connection.Relationship.Direction = "incoming"
				}

				nodeWithConn.Connections = append(nodeWithConn.Connections, connection)
			}

			network = append(network, nodeWithConn)
		}

		return network, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
