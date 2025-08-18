package models

// EntityNetwork represents a network of connected entities and their relationships
type EntityNetwork struct {
	CentralEntity   *ExtractedEntity         `json:"centralEntity"`
	RelatedEntities []*ExtractedEntity       `json:"relatedEntities"`
	Relationships   []*ExtractedRelationship `json:"relationships"`
}
