package sitegenerator

import (
	"io"
	"text/template"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
)

var (
	conceptPageTemplate = template.Must(template.New("concept").Parse(`
### Concept Map: [{{.ConceptMap.Title}}](../../{{.ConceptMap.Slug}}/summary.md)
# Concept: {{.Concept.Label}}
{{.Concept.Description}}

## Diagram
![{{.Concept.Label}}]({{.Diagram}})

## Related Concepts {{ range .RelatedConcepts }}{{ $c := . }}
### [{{.Label}}]({{.Key}}.md)
{{.Description}}{{ range $.ConceptMap.Propositions.InvolvingConcepts . }}
- {{ if (eq .Left.Key $c.Key) }}{{.Left.Label}} {{.Predicate}} [{{.Right.Label}}]({{.Right.Key}}.md){{else}}[{{.Left.Label}}]({{.Left.Key}}.md) {{.Predicate}} {{.Right.Label}}{{ end }}{{ end }}
{{ end }}
`))
)

type conceptPageTemplateData struct {
	ConceptMap      *conceptmap.ConceptMap
	Diagram         string
	Concept         *conceptmap.Concept
	RelatedConcepts []*conceptmap.Concept
}

func NewConceptPageTemplate(conceptMap *conceptmap.ConceptMap, concept *conceptmap.Concept, ph *FilePathHelper) PageTemplate {
	return PageTemplateFunc(func(w io.Writer) error {
		return conceptPageTemplate.Execute(w, &conceptPageTemplateData{
			ConceptMap:      conceptMap,
			Concept:         concept,
			Diagram:         ph.ConceptImageFile(conceptMap, concept),
			RelatedConcepts: conceptMap.ConceptsRelatedTo(concept),
		})
	})
}
