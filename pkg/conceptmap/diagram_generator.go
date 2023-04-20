package conceptmap

import "context"

// DiagramGenerator generates concept map diagrams
type DiagramGenerator interface {
	GenerateConceptMapSVG(ctx context.Context, cmap *ConceptMap, file string) error
	GenerateSingleConceptSVG(ctx context.Context, cmap *ConceptMap, concept *Concept, file string) error
}
