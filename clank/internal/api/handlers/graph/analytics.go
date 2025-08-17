package graph

import (
	"clank/internal/db"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// GetCorruptionScoreHandler calculates a corruption score for a node based on its relationships
func GetCorruptionScoreHandler(c *gin.Context) {
	nodeID := c.Param("nodeId")

	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)-[r]-(m)
			WHERE ID(n) = $nodeId
			WITH n, type(r) as relType, count(r) as relCount
			RETURN n.name as name,
				   collect({type: relType, count: relCount}) as relationships,
				   sum(
					   CASE relType 
						   WHEN 'DONATED_TO' THEN relCount * 2
						   WHEN 'INVESTIGATED_FOR' THEN relCount * 3
						   WHEN 'CONVICTED_OF' THEN relCount * 5
						   ELSE relCount
					   END
				   ) as corruptionScore
		`
		params := map[string]interface{}{
			"nodeId": nodeID,
		}

		result, err := tx.Run(query, params)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("node not found")
		}

		record := result.Record()
		return map[string]interface{}{
			"name":            record.Values[0],
			"relationships":   record.Values[1],
			"corruptionScore": record.Values[2],
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

// GetEntityConnectionsHandler analyzes connections between different types of entities
func GetEntityConnectionsHandler(c *gin.Context) {
	nodeID := c.Param("nodeId")
	depth := c.DefaultQuery("depth", "3")

	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH path = (n)-[*1..%s]-(m)
			WHERE ID(n) = $nodeId
			WITH DISTINCT m, 
				 [(m)-[r]-(o) | type(r)] as relationTypes,
				 [(m)-[r]-(o) | labels(o)[0]] as connectedTypes
			RETURN m.name as name,
				   labels(m)[0] as type,
				   relationTypes,
				   connectedTypes,
				   size(relationTypes) as connectionCount
			ORDER BY connectionCount DESC
		`
		params := map[string]interface{}{
			"nodeId": nodeID,
		}

		result, err := tx.Run(fmt.Sprintf(query, depth), params)
		if err != nil {
			return nil, err
		}

		var connections []map[string]interface{}
		for result.Next() {
			record := result.Record()
			connections = append(connections, map[string]interface{}{
				"name":            record.Values[0],
				"type":            record.Values[1],
				"relationTypes":   record.Values[2],
				"connectedTypes":  record.Values[3],
				"connectionCount": record.Values[4],
			})
		}

		if len(connections) == 0 {
			return nil, fmt.Errorf("no connections found")
		}

		return connections, nil
	})

	if err != nil {
		if err.Error() == "no connections found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTimelineHandler generates a timeline of events
func GetTimelineHandler(c *gin.Context) {
	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)-[r]-(m)
			WHERE exists(r.date)
			RETURN r.date as date,
				   type(r) as eventType,
				   n.name as source,
				   m.name as target,
				   r.amount as amount
			ORDER BY r.date DESC
		`

		result, err := tx.Run(query, nil)
		if err != nil {
			return nil, err
		}

		var events []map[string]interface{}
		for result.Next() {
			record := result.Record()
			events = append(events, map[string]interface{}{
				"date":      record.Values[0],
				"eventType": record.Values[1],
				"source":    record.Values[2],
				"target":    record.Values[3],
				"amount":    record.Values[4],
			})
		}

		if len(events) == 0 {
			return nil, fmt.Errorf("no events found")
		}

		return events, nil
	})

	if err != nil {
		if err.Error() == "no events found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetNetworkStatsHandler provides statistics about the network
func GetNetworkStatsHandler(c *gin.Context) {
	result, err := db.ExecuteRead(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (n)
			WITH count(n) as totalNodes,
				 collect(distinct labels(n)[0]) as nodeTypes
			MATCH ()-[r]->()
			WITH totalNodes, nodeTypes,
				 count(r) as totalRelationships,
				 collect(distinct type(r)) as relationshipTypes
			RETURN {
				totalNodes: totalNodes,
				totalRelationships: totalRelationships,
				nodeTypes: nodeTypes,
				relationshipTypes: relationshipTypes
			} as stats
		`

		result, err := tx.Run(query, nil)
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("failed to get statistics")
		}

		record := result.Record()
		return record.Values[0], nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
