package conceptmap

import (
	"strings"

	"github.com/gosimple/slug"
)

type Predicate string

// Concept is a node in the concept map
type Concept struct {
	Label       string
	Description string
}

// Key is normalised key of the concept
func (c *Concept) Key() string {
	return slug.Make(c.Label)
}

// Proposition is a phrase consisting of two concepts joined by a predicate
type Proposition struct {
	Left      *Concept
	Right     *Concept
	Predicate Predicate
}

type PropositionFilter func(*Proposition) bool

type PropositionList []*Proposition

func (ps PropositionList) Where(fn PropositionFilter) PropositionList {
	output := []*Proposition{}

	for _, p := range ps {
		if fn(p) {
			output = append(output, p)
		}
	}

	return output
}

func (ps PropositionList) InvolvingConcept(c *Concept) PropositionList {
	return ps.Where(func(p *Proposition) bool {
		return p.Left.Key() == c.Key() || p.Right.Key() == c.Key()
	})
}

func (p *Proposition) String() string {
	return strings.Join([]string{p.Left.Label, string(p.Predicate), p.Right.Label}, " ")
}
