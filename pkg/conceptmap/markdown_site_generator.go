package conceptmap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
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

type indexPageTemplateData struct {
	ConceptMaps []*ConceptMap
}

func NewIndexPageTemplateData(conceptMaps []*ConceptMap) *indexPageTemplateData {
	return &indexPageTemplateData{
		ConceptMaps: conceptMaps,
	}
}

type conceptMapSummaryPageTemplateData struct {
	Diagram    string
	ConceptMap *ConceptMap
}

func NewConceptMapSummaryPageTemplateData(conceptMap *ConceptMap) *conceptMapSummaryPageTemplateData {
	return &conceptMapSummaryPageTemplateData{
		Diagram:    NewFilePathHelper("../../").ConceptMapSummaryImageFile(conceptMap),
		ConceptMap: conceptMap,
	}
}

type conceptMapDetailPageTemplateData struct {
	Diagram    string
	ConceptMap *ConceptMap
}

func NewConceptMapDetailPageTemplateData(conceptMap *ConceptMap) *conceptMapSummaryPageTemplateData {
	return &conceptMapSummaryPageTemplateData{
		Diagram:    NewFilePathHelper("../../").ConceptMapDetailImageFile(conceptMap),
		ConceptMap: conceptMap,
	}
}

type conceptPageTemplateData struct {
	ConceptMap      *ConceptMap
	Diagram         string
	Concept         *Concept
	RelatedConcepts []*Concept
}

func NewConceptPageTemplateData(conceptMap *ConceptMap, concept *Concept) *conceptPageTemplateData {
	return &conceptPageTemplateData{
		ConceptMap:      conceptMap,
		Concept:         concept,
		Diagram:         NewFilePathHelper("../../../").ConceptImageFile(conceptMap, concept),
		RelatedConcepts: conceptMap.ConceptsRelatedTo(concept),
	}
}

type SiteGenerator interface {
	GenerateSite(ctx context.Context, cmaps []*ConceptMap) error
}

type MarkdownSiteGenerator struct {
	diagramGenerator DiagramGenerator
	filePathHelper   *FilePathHelper
}

func NewMarkdownSiteGenerator(dg DiagramGenerator, outputDir string) SiteGenerator {
	return &MarkdownSiteGenerator{
		diagramGenerator: dg,
		filePathHelper:   NewFilePathHelper(outputDir),
	}
}

func (sg *MarkdownSiteGenerator) GenerateSite(ctx context.Context, cmaps []*ConceptMap) error {
	if err := os.MkdirAll(sg.filePathHelper.ImageDir(), os.ModePerm); err != nil {
		return err
	}

	if err := sg.renderTemplateToFile(
		sg.filePathHelper.IndexMarkdownFile(),
		indexPageTemplate,
		NewIndexPageTemplateData(cmaps)); err != nil {
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

func (sg *MarkdownSiteGenerator) generateConceptMapSummaryPage(ctx context.Context, cmap *ConceptMap) error {
	diagramFile := sg.filePathHelper.ConceptMapSummaryImageFile(cmap)

	if err := sg.diagramGenerator.GenerateConceptMapSummarySVG(ctx, cmap, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMapSummaryMarkdownFile(cmap),
		conceptMapSummaryPageTemplate,
		NewConceptMapSummaryPageTemplateData(cmap))
}

func (sg *MarkdownSiteGenerator) generateConceptMapDetailPage(ctx context.Context, cmap *ConceptMap) error {
	diagramFile := sg.filePathHelper.ConceptMapDetailImageFile(cmap)

	if err := sg.diagramGenerator.GenerateConceptMapDetailSVG(ctx, cmap, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMapDetailMarkdownFile(cmap),
		conceptMapDetailPageTemplate,
		NewConceptMapDetailPageTemplateData(cmap))
}

func (sg *MarkdownSiteGenerator) generateConceptPage(ctx context.Context, cmap *ConceptMap, concept *Concept) error {
	diagramFile := sg.filePathHelper.ConceptImageFile(cmap, concept)

	if err := sg.diagramGenerator.GenerateSingleConceptSVG(ctx, cmap, concept, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMarkdownFile(cmap, concept),
		conceptPageTemplate,
		NewConceptPageTemplateData(cmap, concept))
}

func (sg *MarkdownSiteGenerator) renderTemplateToFile(file string, tpl *template.Template, data any) error {
	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	return tpl.Execute(f, data)
}

type FilePathHelper struct {
	BaseDir string
}

func NewFilePathHelper(baseDir string) *FilePathHelper {
	return &FilePathHelper{
		BaseDir: baseDir,
	}
}

func (h *FilePathHelper) ImageDir() string {
	return filepath.Join(h.BaseDir, "images")
}

func (h *FilePathHelper) ConceptMapSummaryImageFile(conceptMap *ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		conceptMap.Slug(),
		"images",
		fmt.Sprintf("%s-summary.svg", conceptMap.Slug()))
}

func (h *FilePathHelper) ConceptMapDetailImageFile(conceptMap *ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		conceptMap.Slug(),
		"images",
		fmt.Sprintf("%s-detail.svg", conceptMap.Slug()))
}

func (h *FilePathHelper) IndexMarkdownFile() string {
	return filepath.Join(h.BaseDir, "index.md")
}

func (h *FilePathHelper) ConceptMapSummaryMarkdownFile(conceptMap *ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s/summary.md", conceptMap.Slug()))
}

func (h *FilePathHelper) ConceptMapDetailMarkdownFile(conceptMap *ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s/detail.md", conceptMap.Slug()))
}

func (h *FilePathHelper) ConceptImageFile(conceptMap *ConceptMap, concept *Concept) string {
	return filepath.Join(
		h.BaseDir,
		conceptMap.Slug(),
		"images",
		fmt.Sprintf("%s.svg", concept.Key()))
}

func (h *FilePathHelper) ConceptMarkdownFile(conceptMap *ConceptMap, concept *Concept) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s/concepts/%s.md", conceptMap.Slug(), concept.Key()))
}
