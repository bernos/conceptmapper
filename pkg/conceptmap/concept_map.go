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

// HasKeyConcepts returns true if one or more concepts in the concept map is marked as a key concept
func (m *ConceptMap) HasKeyConcepts() bool {
	for _, c := range m.Concepts {
		if c.IsKeyConcept {
			return true
		}
	}
	return false
}

func (m *ConceptMap) KeyConcepts() []*Concept {
	output := []*Concept{}

	for _, c := range m.Concepts {
		if c.IsKeyConcept {
			output = append(output, c)
		}
	}

	return output
}

// ConceptsRelatedTo returns all concepts that are related to c via a Proposition
func (m *ConceptMap) ConceptsRelatedTo(concepts ...*Concept) []*Concept {
	output := []*Concept{}
	rm := map[string]*Concept{}

	for _, c := range concepts {
		for _, p := range m.Propositions {
			if p.Left.Key() == c.Key() {
				rm[p.Right.Key()] = p.Right
			}

			if p.Right.Key() == c.Key() {
				rm[p.Left.Key()] = p.Left
			}
		}
	}

	for _, v := range rm {
		output = append(output, v)
	}

	return output
}
