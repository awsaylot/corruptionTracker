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

// CreateRelationship creates a new relationship between two nodes
func CreateRelationship(c *gin.Context) {
	var rel models.Relationship
	if err := c.ShouldBindJSON(&rel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add timestamp to properties
	if rel.Props == nil {
		rel.Props = make(map[string]any)
	}
	rel.Props["created_at"] = time.Now()

	result, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (from), (to)
			WHERE ID(from) = $fromId AND ID(to) = $toId
			CREATE (from)-[r:%s $props]->(to)
			RETURN r
		`
		params := map[string]interface{}{
			"fromId": rel.FromID,
			"toId":   rel.ToID,
			"props":  rel.Props,
		}

		result, err := tx.Run(fmt.Sprintf(query, rel.Type), params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("relationship not created")
		}

		record := result.Record()
		relationship := record.Values[0].(neo4j.Relationship)

		return models.Relationship{
			ID:     fmt.Sprint(relationship.Id),
			FromID: fmt.Sprint(relationship.StartId),
			ToID:   fmt.Sprint(relationship.EndId),
			Type:   relationship.Type,
			Props:  relationship.Props,
		}, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetRelationship returns a specific relationship by ID
func GetRelationship(c *gin.Context) {
	id := c.Param("id")

	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH ()-[r]->()
			WHERE ID(r) = $id
			RETURN r
		`
		params := map[string]interface{}{
			"id": id,
		}

		result, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("relationship not found")
		}

		record := result.Record()
		relationship := record.Values[0].(neo4j.Relationship)

		return models.Relationship{
			ID:     fmt.Sprint(relationship.Id),
			FromID: fmt.Sprint(relationship.StartId),
			ToID:   fmt.Sprint(relationship.EndId),
			Type:   relationship.Type,
			Props:  relationship.Props,
		}, nil
	})

	if err != nil {
		if err.Error() == "relationship not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateRelationship updates a relationship by ID
func UpdateRelationship(c *gin.Context) {
	id := c.Param("id")
	var update models.Relationship
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
			MATCH ()-[r]->()
			WHERE ID(r) = $id
			SET r += $props
			RETURN r
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
			return nil, fmt.Errorf("relationship not found")
		}

		record := result.Record()
		relationship := record.Values[0].(neo4j.Relationship)

		return models.Relationship{
			ID:     fmt.Sprint(relationship.Id),
			FromID: fmt.Sprint(relationship.StartId),
			ToID:   fmt.Sprint(relationship.EndId),
			Type:   relationship.Type,
			Props:  relationship.Props,
		}, nil
	})

	if err != nil {
		if err.Error() == "relationship not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteRelationship deletes a relationship by ID
func DeleteRelationship(c *gin.Context) {
	id := c.Param("id")

	_, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH ()-[r]->()
			WHERE ID(r) = $id
			DELETE r
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

		if summary.Counters().RelationshipsDeleted() == 0 {
			return nil, fmt.Errorf("relationship not found")
		}

		return nil, nil
	})

	if err != nil {
		if err.Error() == "relationship not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
