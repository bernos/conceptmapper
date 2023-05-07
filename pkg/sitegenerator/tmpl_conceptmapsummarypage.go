package sitegenerator

import (
	"io"
	"text/template"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
)

var conceptMapSummaryPageTemplate = template.Must(template.New("concept-map-summary").Parse(`
# Concept Map: {{.ConceptMap.Title}}
{{.ConceptMap.Description}}
{{ if .ConceptMap.HasKeyConcepts }}

> This is a summary of the key concepts in this map. You might also like to [view the map in its entirety](detail.md).

{{ end }}
## Diagram
![{{.ConceptMap.Title}}]({{.Diagram}})

{{ if .ConceptMap.HasKeyConcepts }}
## Concepts {{ range .ConceptMap.KeyConcepts }}{{ $c := . }}
### [{{.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Key}}.md)
{{.Description}}{{ range $.ConceptMap.Propositions.InvolvingConcepts . }}
- {{ if (eq .Left.Key $c.Key) }}{{.Left.Label}} {{.Predicate}} [{{.Right.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Right.Key}}.md){{else}}[{{.Left.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Left.Key}}.md) {{.Predicate}} {{.Right.Label}}{{ end }}{{ end }}
{{ end }}
{{ else }}
## Concepts {{ range .ConceptMap.Concepts }}{{ $c := . }}
### [{{.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Key}}.md)
{{.Description}}{{ range $.ConceptMap.Propositions.InvolvingConcepts . }}
- {{ if (eq .Left.Key $c.Key) }}{{.Left.Label}} {{.Predicate}} [{{.Right.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Right.Key}}.md){{else}}[{{.Left.Label}}](../{{$.ConceptMap.Slug}}/concepts/{{.Left.Key}}.md) {{.Predicate}} {{.Right.Label}}{{ end }}{{ end }}
{{ end }}
{{ end}}
`))

type conceptMapSummaryPageTemplateData struct {
	Diagram    string
	ConceptMap *conceptmap.ConceptMap
}

func NewConceptMapSummaryPageTemplate(conceptMap *conceptmap.ConceptMap) PageTemplate {
	return PageTemplateFunc(func(w io.Writer) error {
		return conceptMapSummaryPageTemplate.Execute(w, &conceptMapSummaryPageTemplateData{
			Diagram:    NewFilePathHelper("../../").ConceptMapSummaryImageFile(conceptMap),
			ConceptMap: conceptMap,
		})
	})
}
