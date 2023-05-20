package sitegenerator

import (
	"io"
	"text/template"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
)

var conceptMapDetailPageTemplate = template.Must(template.New("concept-map-detail").Parse(`
# Concept Map: {{.ConceptMap.Title}}
{{.ConceptMap.Description}}

> This is a detailed view of this map. You might also like to [view a summary of the key concepts](summary.md).

## Diagram
![{{.ConceptMap.Title}}]({{.Diagram}})

## Concepts {{ range .ConceptMap.Concepts }}{{ $c := . }}
### [{{.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Key}}.md)
{{.Description}}{{ range $.ConceptMap.Propositions.InvolvingConcepts . }}
- {{ if (eq .Left.Key $c.Key) }}{{.Left.Label}} {{.Predicate}} [{{.Right.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Right.Key}}.md){{else}}[{{.Left.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Left.Key}}.md) {{.Predicate}} {{.Right.Label}}{{ end }}{{ end }}
{{ end }}
`))

type conceptMapDetailPageTemplateData struct {
	Diagram    string
	ConceptMap *conceptmap.ConceptMap
}

func NewConceptMapDetailPageTemplate(conceptMap *conceptmap.ConceptMap, ph *FilePathHelper) PageTemplate {
	return PageTemplateFunc(func(w io.Writer) error {
		return conceptMapDetailPageTemplate.Execute(w, &conceptMapSummaryPageTemplateData{
			Diagram:    ph.ConceptMapDetailImageFile(conceptMap),
			ConceptMap: conceptMap,
		})
	})

}
