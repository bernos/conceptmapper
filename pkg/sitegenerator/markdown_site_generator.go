package sitegenerator

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
)

var (
	indexPageTemplate = template.Must(template.New("index").Parse(`
# Concept Maps{{ range .ConceptMaps }}
## [{{.Title}}](./{{.Slug}}/summary.md)
{{.Description}}
{{end}}
`))

	conceptMapSummaryPageTemplate = template.Must(template.New("concept-map-summary").Parse(`
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

	conceptMapDetailPageTemplate = template.Must(template.New("concept-map-detail").Parse(`
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

type PageTemplate interface {
	Render(io.Writer) error
}

type PageTemplateFunc func(io.Writer) error

func (fn PageTemplateFunc) Render(w io.Writer) error {
	return fn(w)
}

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

type conceptMapDetailPageTemplateData struct {
	Diagram    string
	ConceptMap *conceptmap.ConceptMap
}

func NewConceptMapDetailPageTemplate(conceptMap *conceptmap.ConceptMap) PageTemplate {
	return PageTemplateFunc(func(w io.Writer) error {
		return conceptMapDetailPageTemplate.Execute(w, &conceptMapSummaryPageTemplateData{
			Diagram:    NewFilePathHelper("../../").ConceptMapDetailImageFile(conceptMap),
			ConceptMap: conceptMap,
		})
	})

}

type conceptPageTemplateData struct {
	ConceptMap      *conceptmap.ConceptMap
	Diagram         string
	Concept         *conceptmap.Concept
	RelatedConcepts []*conceptmap.Concept
}

func NewConceptPageTemplate(conceptMap *conceptmap.ConceptMap, concept *conceptmap.Concept) PageTemplate {
	return PageTemplateFunc(func(w io.Writer) error {
		return conceptPageTemplate.Execute(w, &conceptPageTemplateData{
			ConceptMap:      conceptMap,
			Concept:         concept,
			Diagram:         NewFilePathHelper("../../../").ConceptImageFile(conceptMap, concept),
			RelatedConcepts: conceptMap.ConceptsRelatedTo(concept),
		})
	})
}

type MarkdownSiteGenerator struct {
	diagramGenerator conceptmap.DiagramGenerator
	filePathHelper   *FilePathHelper
}

func NewMarkdownSiteGenerator(dg conceptmap.DiagramGenerator, outputDir string) *MarkdownSiteGenerator {
	return &MarkdownSiteGenerator{
		diagramGenerator: dg,
		filePathHelper:   NewFilePathHelper(outputDir),
	}
}

func (sg *MarkdownSiteGenerator) GenerateSite(ctx context.Context, cmaps []*conceptmap.ConceptMap) error {

	if err := sg.renderTemplateToFile(
		sg.filePathHelper.IndexMarkdownFile(),
		NewIndexPageTemplate(cmaps)); err != nil {
		return err
	}

	for _, cmap := range cmaps {
		if err := sg.generateConceptMapSummaryPage(ctx, cmap); err != nil {
			return err
		}

		if cmap.HasKeyConcepts() {
			if err := sg.generateConceptMapDetailPage(ctx, cmap); err != nil {
				return err
			}
		}

		for _, concept := range cmap.Concepts {
			if err := sg.generateConceptPage(ctx, cmap, concept); err != nil {
				return err
			}
		}
	}

	return nil
}

func (sg *MarkdownSiteGenerator) generateConceptMapSummaryPage(ctx context.Context, cmap *conceptmap.ConceptMap) error {
	diagramFile := sg.filePathHelper.ConceptMapSummaryImageFile(cmap)

	if err := sg.diagramGenerator.GenerateConceptMapSummarySVG(ctx, cmap, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMapSummaryMarkdownFile(cmap),
		NewConceptMapSummaryPageTemplate(cmap))
}

func (sg *MarkdownSiteGenerator) generateConceptMapDetailPage(ctx context.Context, cmap *conceptmap.ConceptMap) error {
	diagramFile := sg.filePathHelper.ConceptMapDetailImageFile(cmap)

	if err := sg.diagramGenerator.GenerateConceptMapDetailSVG(ctx, cmap, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMapDetailMarkdownFile(cmap),
		NewConceptMapDetailPageTemplate(cmap))
}

func (sg *MarkdownSiteGenerator) generateConceptPage(ctx context.Context, cmap *conceptmap.ConceptMap, concept *conceptmap.Concept) error {
	diagramFile := sg.filePathHelper.ConceptImageFile(cmap, concept)

	if err := sg.diagramGenerator.GenerateSingleConceptSVG(ctx, cmap, concept, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMarkdownFile(cmap, concept),
		NewConceptPageTemplate(cmap, concept))
}

func (sg *MarkdownSiteGenerator) renderTemplateToFile(file string, tpl PageTemplate) error {
	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return tpl.Render(f)
}

type FilePathHelper struct {
	BaseDir string
}

func NewFilePathHelper(baseDir string) *FilePathHelper {
	return &FilePathHelper{
		BaseDir: baseDir,
	}
}

func (h *FilePathHelper) ConceptMapSummaryImageFile(conceptMap *conceptmap.ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		conceptMap.Slug(),
		"images",
		fmt.Sprintf("%s-summary.svg", conceptMap.Slug()))
}

func (h *FilePathHelper) ConceptMapDetailImageFile(conceptMap *conceptmap.ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		conceptMap.Slug(),
		"images",
		fmt.Sprintf("%s-detail.svg", conceptMap.Slug()))
}

func (h *FilePathHelper) IndexMarkdownFile() string {
	return filepath.Join(h.BaseDir, "index.md")
}

func (h *FilePathHelper) ConceptMapSummaryMarkdownFile(conceptMap *conceptmap.ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s/summary.md", conceptMap.Slug()))
}

func (h *FilePathHelper) ConceptMapDetailMarkdownFile(conceptMap *conceptmap.ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s/detail.md", conceptMap.Slug()))
}

func (h *FilePathHelper) ConceptImageFile(conceptMap *conceptmap.ConceptMap, concept *conceptmap.Concept) string {
	return filepath.Join(
		h.BaseDir,
		conceptMap.Slug(),
		"images",
		fmt.Sprintf("%s.svg", concept.Key()))
}

func (h *FilePathHelper) ConceptMarkdownFile(conceptMap *conceptmap.ConceptMap, concept *conceptmap.Concept) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s/concepts/%s.md", conceptMap.Slug(), concept.Key()))
}
