package sitegenerator

import (
	"context"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
)

// DiagramGenerator generates concept map diagrams
type DiagramGenerator interface {
	GenerateConceptMapSummarySVG(ctx context.Context, cmap *conceptmap.ConceptMap, file string) error
	GenerateConceptMapDetailSVG(ctx context.Context, cmap *conceptmap.ConceptMap, file string) error
	GenerateSingleConceptSVG(ctx context.Context, cmap *conceptmap.ConceptMap, concept *conceptmap.Concept, file string) error
}
