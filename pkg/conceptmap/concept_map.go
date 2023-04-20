package conceptmap

import (
	"github.com/gosimple/slug"
)

// ConceptMap is the main datastructure that represents our concept map
type ConceptMap struct {
	Title        string
	Description  string
	Propositions PropositionList
	Concepts     []*Concept
}

// Slug is the slugified version of Map.Title
func (m *ConceptMap) Slug() string {
	return slug.Make(m.Title)
}

// ConceptsRelatedTo returns all concepts that are related to c via a Proposition
func (m *ConceptMap) ConceptsRelatedTo(c *Concept) []*Concept {
	output := []*Concept{}
	rm := map[string]*Concept{}

	for _, p := range m.Propositions {
		if p.Left.Key() == c.Key() {
			rm[p.Right.Key()] = p.Right
		}

		if p.Right.Key() == c.Key() {
			rm[p.Left.Key()] = p.Left
		}
	}

	for _, v := range rm {
		output = append(output, v)
	}

	return output
}
