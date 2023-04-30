package conceptmap

import (
	"strings"
)

type Predicate string

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

func (ps PropositionList) InvolvingConcepts(cs ...*Concept) PropositionList {
	return ps.Where(func(p *Proposition) bool {
		for _, c := range cs {
			if p.Left.Key() == c.Key() || p.Right.Key() == c.Key() {
				return true
			}
		}
		return false
	})
}

func (ps PropositionList) ConnectingConcepts(cs ...*Concept) PropositionList {
	return ps.Where(func(p *Proposition) bool {
		l := false
		r := false
		for _, c := range cs {
			if p.Left.Key() == c.Key() {
				l = true
			}

			if p.Right.Key() == c.Key() {
				r = true
			}

			if l && r {
				return true
			}
		}
		return false
	})
}

func (p *Proposition) String() string {
	return strings.Join([]string{p.Left.Label, string(p.Predicate), p.Right.Label}, " ")
}
