package conceptmap

import "context"

// DiagramGenerator generates concept map diagrams
type DiagramGenerator interface {
	GenerateConceptMapSummarySVG(ctx context.Context, cmap *ConceptMap, file string) error
	GenerateConceptMapDetailSVG(ctx context.Context, cmap *ConceptMap, file string) error
	GenerateSingleConceptSVG(ctx context.Context, cmap *ConceptMap, concept *Concept, file string) error
}
