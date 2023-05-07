package diagrams

import (
	"fmt"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2oracle"
)

type D2GraphModifier func(*d2graph.Graph) (*d2graph.Graph, error)

func emphasiseConceptWithKey(key string) D2GraphModifier {
	return func(g *d2graph.Graph) (*d2graph.Graph, error) {
		s := "true"
		return d2oracle.Set(g, fmt.Sprintf("%s.style.underline", key), nil, &s)
	}
}

func addLinksToConcepts(cmap *conceptmap.ConceptMap, concepts []*conceptmap.Concept) D2GraphModifier {
	return func(g *d2graph.Graph) (*d2graph.Graph, error) {
		for _, concept := range concepts {
			// link := fmt.Sprintf("http://localhost:8080/%s_%s/", cmap.Slug(), concept.Key())
			link := "https://google.com"

			g, err := d2oracle.Set(g, fmt.Sprintf("%s.link", concept.Key()), nil, &link)
			if err != nil {
				return g, err
			}
		}
		return g, nil
	}
}
