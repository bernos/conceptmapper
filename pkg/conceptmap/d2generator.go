package conceptmap

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gosimple/slug"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

type D2DiagramGeneratorOption func(*D2DiagramGenerator)
type D2GraphModifier func(*d2graph.Graph) (*d2graph.Graph, error)
type D2RulerFactory func() (*textmeasure.Ruler, error)

func WithD2GraphModifier(mod D2GraphModifier) D2DiagramGeneratorOption {
	return func(d *D2DiagramGenerator) {
		d.graphModifiers = append(d.graphModifiers, mod)
	}
}

func WithD2RulerFactory(fn D2RulerFactory) D2DiagramGeneratorOption {
	return func(d *D2DiagramGenerator) {
		d.rulerFactory = fn
	}
}

type D2DiagramGenerator struct {
	ruler          *textmeasure.Ruler
	rulerFactory   D2RulerFactory
	graphModifiers []D2GraphModifier
}

func NewD2DiagramGenerator(opts ...D2DiagramGeneratorOption) *D2DiagramGenerator {
	d := &D2DiagramGenerator{
		rulerFactory:   defaultRulerFactory,
		graphModifiers: []D2GraphModifier{},
	}

	for _, o := range opts {
		o(d)
	}

	return d
}

func (d *D2DiagramGenerator) D2Script(ctx context.Context, propositions []*Proposition, modifiers ...D2GraphModifier) (string, error) {
	var err error

	_, graph, err := d2lib.Compile(ctx, "", nil)
	if err != nil {
		return "", nil
	}

	// Map concept slugs to descriptions
	concepts := map[string]string{}

	// Keeps track of which left concepts we've already joined to their predicates
	leftConceptToPredicateEdges := map[string]int{}

	for _, proposition := range propositions {

		predicate := (string)(proposition.Predicate)
		predicateKey := slug.Make(strings.Join([]string{proposition.Left.Key(), predicate, proposition.Right.Key()}, " "))
		predicateKey = slug.Make(strings.Join([]string{proposition.Left.Key(), predicate}, " "))

		concepts[proposition.Left.Key()] = proposition.Left.Label
		concepts[proposition.Right.Key()] = proposition.Right.Label

		// Only draw edges from the left concept to identical predicates once
		_, ok := leftConceptToPredicateEdges[predicateKey]
		if !ok {

			italic := "false"
			bold := "false"
			text := "text"

			// Style it!
			graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.shape", predicateKey), nil, &text)
			if err != nil {
				return "", err
			}

			graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.style.italic", predicateKey), nil, &italic)
			if err != nil {
				return "", err
			}

			graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.style.bold", predicateKey), nil, &bold)
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

		// Predicate -> Right Concept
		graph, _, err = d2oracle.Create(graph, fmt.Sprintf("%s -> %s", predicateKey, proposition.Right.Key()))
		if err != nil {
			return "", err
		}
	}

	for _, mod := range modifiers {
		graph, err = mod(graph)
		if err != nil {
			return "", err
		}
	}

	// Now, add each concept to the graph and add its label
	for k, v := range concepts {
		graph, err = d2oracle.Set(graph, fmt.Sprintf("%s.label", k), nil, &v)
		if err != nil {
			return "", err
		}
	}

	return d2format.Format(graph.AST), nil
}

func (d *D2DiagramGenerator) GenerateConceptMapSVG(ctx context.Context, cmap *ConceptMap, file string) error {
	script, err := d.D2Script(ctx, cmap.Propositions)
	if err != nil {
		return err
	}

	return d.generateSVGFileFromScript(ctx, script, file)
}

func (d *D2DiagramGenerator) GenerateSingleConceptSVG(ctx context.Context, cmap *ConceptMap, concept *Concept, file string) error {
	filtered := cmap.Propositions.InvolvingConcept(concept)

	script, err := d.D2Script(ctx, filtered, emphasiseConceptWithKey(concept.Key()))
	if err != nil {
		return err
	}

	return d.generateSVGFileFromScript(ctx, script, file)
}

func (d *D2DiagramGenerator) generateSVGFileFromScript(ctx context.Context, script string, file string) error {

	ruler, err := d.getRuler()
	if err != nil {
		return err
	}

	diagram, _, _ := d2lib.Compile(ctx, script, &d2lib.CompileOptions{
		Layout: d2dagrelayout.DefaultLayout,
		Ruler:  ruler,
	})

	out, err := d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad: d2svg.DEFAULT_PADDING,
	})

	if err != nil {
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

func defaultRulerFactory() (*textmeasure.Ruler, error) {
	return textmeasure.NewRuler()
}

func emphasiseConceptWithKey(key string) D2GraphModifier {
	return func(g *d2graph.Graph) (*d2graph.Graph, error) {
		s := "true"
		return d2oracle.Set(g, fmt.Sprintf("%s.style.underline", key), nil, &s)
	}
}
