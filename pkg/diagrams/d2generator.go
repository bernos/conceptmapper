package diagrams

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
	"github.com/gosimple/slug"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

type D2DiagramGenerator struct {
	direction      Direction
	ruler          *textmeasure.Ruler
	rulerFactory   D2RulerFactory
	graphModifiers []D2GraphModifier
}

func NewD2DiagramGenerator(opts ...D2DiagramGeneratorOption) *D2DiagramGenerator {
	d := &D2DiagramGenerator{
		direction:      DirectionDown,
		rulerFactory:   defaultRulerFactory,
		graphModifiers: []D2GraphModifier{},
	}

	for _, o := range opts {
		o(d)
	}

	return d
}

func (d *D2DiagramGenerator) D2Script(ctx context.Context, propositions []*conceptmap.Proposition, modifiers ...D2GraphModifier) (string, error) {
	var err error

	script := `direction: %s
classes: {
	concept: {
		height: 32
	}		
	conceptRounded: {
		height: 32
		style: {
			border-radius: 8
		}
	}
	predicate: {
		shape: text
		height: 32
		style: {
			italic: true
		}
	}
}`

	_, graph, err := d2lib.Compile(ctx, fmt.Sprintf(script, d.direction), nil)
	if err != nil {
		return "", nil
	}

	// Keeps track of which left concepts we've already joined to their predicates
	leftConceptToPredicateEdges := map[string]int{}

	predicateClass := "predicate"
	defaultConceptClass := "concept"

	appendConcept := d.distinctConceptAppenderFunc()

	for _, mod := range modifiers {
		graph, err = mod(graph)
		if err != nil {
			return "", err
		}
	}

	for _, proposition := range propositions {

		predicate := (string)(proposition.Predicate)
		predicateKey := slug.Make(strings.Join([]string{proposition.Left.Key(), predicate}, " "))
		leftClass := defaultConceptClass
		rightClass := defaultConceptClass

		// Draw concepts
		if predicate == "is a" || predicate == "is an" {
			leftClass = "conceptRounded"
		}

		// Left concept
		graph, err = appendConcept(graph, proposition.Left, leftClass)
		if err != nil {
			return "", err
		}

		// Right concept
		graph, err = appendConcept(graph, proposition.Right, rightClass)
		if err != nil {
			return "", err
		}

		// Predicate -> Right Concept
		graph, _, err = d2oracle.Create(graph, fmt.Sprintf("%s -> %s", predicateKey, proposition.Right.Key()))
		if err != nil {
			return "", err
		}

		// Only draw edges from the left concept to identical predicates once
		_, ok := leftConceptToPredicateEdges[predicateKey]
		if !ok {

			// Style it!
			graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.class", predicateKey), nil, &predicateClass)
			if err != nil {
				return "", err
			}

			// Label must go last
			graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.label", predicateKey), nil, &predicate)
			if err != nil {
				return "", err
			}

			// Left Concept -> Predicate
			graph, _, err = d2oracle.Create(graph, fmt.Sprintf("%s -> %s", proposition.Left.Key(), predicateKey))
			if err != nil {
				return "", err
			}

			leftConceptToPredicateEdges[predicateKey] = 1
		}

	}

	return d2format.Format(graph.AST), nil
}

func (d *D2DiagramGenerator) GenerateConceptMapSummarySVG(ctx context.Context, cmap *conceptmap.ConceptMap, file string) error {
	propositions := cmap.Propositions

	if cmap.HasKeyConcepts() {
		propositions = cmap.Propositions.ConnectingConcepts(cmap.KeyConcepts()...)
	}

	script, err := d.D2Script(ctx, propositions)
	if err != nil {
		return err
	}

	return d.generateSVGFileFromScript(ctx, script, file)
}

func (d *D2DiagramGenerator) GenerateConceptMapDetailSVG(ctx context.Context, cmap *conceptmap.ConceptMap, file string) error {
	script, err := d.D2Script(ctx, cmap.Propositions)
	if err != nil {
		return err
	}

	return d.generateSVGFileFromScript(ctx, script, file)
}

func (d *D2DiagramGenerator) GenerateSingleConceptSVG(ctx context.Context, cmap *conceptmap.ConceptMap, concept *conceptmap.Concept, file string) error {
	filtered := cmap.Propositions.InvolvingConcepts(concept)

	script, err := d.D2Script(ctx, filtered, emphasiseConceptWithKey(concept.Key()))
	if err != nil {
		return err
	}

	return d.generateSVGFileFromScript(ctx, script, file)
}

// distinctConceptAppenderFunc returns a function that appends a concept to a graph
// exactly once. Calling the function multiple times with the same concept will only
// append the concept once. Concepts with the same Key are considered equivalent
func (d *D2DiagramGenerator) distinctConceptAppenderFunc() func(*d2graph.Graph, *conceptmap.Concept, string) (*d2graph.Graph, error) {

	// Keeps track of which concepts we've added to the diagram, so that we can avoid
	// adding them more than once
	appendedConcepts := map[string]int{}

	return func(graph *d2graph.Graph, concept *conceptmap.Concept, class string) (*d2graph.Graph, error) {
		_, alreadyAdded := appendedConcepts[concept.Key()]

		if !alreadyAdded {
			var err error

			graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.class", concept.Key()), nil, &class)
			if err != nil {
				return graph, err
			}

			graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.label", concept.Key()), nil, &concept.Label)
			if err != nil {
				return graph, err
			}

			appendedConcepts[concept.Key()] = 1
		}

		return graph, nil
	}
}

func (d *D2DiagramGenerator) generateSVGFileFromScript(ctx context.Context, script string, file string) error {

	ruler, err := d.getRuler()
	if err != nil {
		return err
	}

	layout := func(ctx context.Context, g *d2graph.Graph) (err error) {
		return d2dagrelayout.Layout(ctx, g, &d2dagrelayout.ConfigurableOpts{
			NodeSep: 30,
			EdgeSep: 10,
		})
	}

	diagram, _, _ := d2lib.Compile(ctx, script, &d2lib.CompileOptions{
		Layout: layout,
		Ruler:  ruler,
	})

	out, err := d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad: d2svg.DEFAULT_PADDING,
	})

	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return err
	}

	return ioutil.WriteFile(file, out, os.ModePerm)
}

func (d *D2DiagramGenerator) getRuler() (*textmeasure.Ruler, error) {
	var err error

	if d.ruler == nil {
		d.ruler, err = d.rulerFactory()
		if err != nil {
			return nil, err
		}
	}

	return d.ruler, nil
}
