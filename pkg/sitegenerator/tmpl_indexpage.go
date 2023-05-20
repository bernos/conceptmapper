package sitegenerator

import (
	"io"
	"text/template"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
)

var indexPageTemplate = template.Must(template.New("index").Parse(`
# Concept Maps{{ range .ConceptMaps }}
## [{{.Title}}](./{{.Slug}}/summary.md)
{{.Description}}
{{end}}
`))

type indexPageTemplateData struct {
	ConceptMaps []*conceptmap.ConceptMap
}

func NewIndexPageTemplate(conceptMaps []*conceptmap.ConceptMap) PageTemplate {
	return PageTemplateFunc(func(w io.Writer) error {
		return indexPageTemplate.Execute(w, &indexPageTemplateData{
			ConceptMaps: conceptMaps,
		})
	})
}
