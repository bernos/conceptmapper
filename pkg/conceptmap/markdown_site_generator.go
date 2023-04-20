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
## [{{.Title}}](./{{.Slug}}.md)
{{.Description}}
{{end}}
`))

	conceptMapPageTemplate = template.Must(template.New("concept-map").Parse(`
# Concept Map: {{.ConceptMap.Title}}
{{.ConceptMap.Description}}

## Diagram
![{{.ConceptMap.Title}}]({{.Diagram}})

## Concepts {{ range .ConceptMap.Concepts }}{{ $c := . }}
### [{{.Label}}](./{{$.ConceptMap.Slug}}_{{.Key}}.md)
{{.Description}}{{ range $.ConceptMap.Propositions.InvolvingConcept . }}
- {{ if (eq .Left.Key $c.Key) }}{{.Left.Label}} {{.Predicate}} [{{.Right.Label}}](./{{$.ConceptMap.Slug}}_{{.Right.Key}}.md){{else}}[{{.Left.Label}}](./{{$.ConceptMap.Slug}}_{{.Left.Key}}.md) {{.Predicate}} {{.Right.Label}}{{ end }}{{ end }}
{{ end }}
`))

	conceptPageTemplate = template.Must(template.New("concept").Parse(`
### Concept Map: [{{.ConceptMap.Title}}](./{{.ConceptMap.Slug}}.md)
# Concept: {{.Concept.Label}}
{{.Concept.Description}}

## Diagram
![{{.Concept.Label}}]({{.Diagram}})

## Related Concepts {{ range .RelatedConcepts }}{{ $c := . }}
### [{{.Label}}](./{{$.ConceptMap.Slug}}_{{.Key}}.md)
{{.Description}}{{ range $.ConceptMap.Propositions.InvolvingConcept . }}
- {{ if (eq .Left.Key $c.Key) }}{{.Left.Label}} {{.Predicate}} [{{.Right.Label}}](./{{$.ConceptMap.Slug}}_{{.Right.Key}}.md){{else}}[{{.Left.Label}}](./{{$.ConceptMap.Slug}}_{{.Left.Key}}.md) {{.Predicate}} {{.Right.Label}}{{ end }}{{ end }}
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

type conceptMapPageTemplateData struct {
	Diagram    string
	ConceptMap *ConceptMap
}

func NewConceptMapPageTemplateData(conceptMap *ConceptMap) *conceptMapPageTemplateData {
	return &conceptMapPageTemplateData{
		Diagram:    NewFilePathHelper("").ConceptMapImageFile(conceptMap),
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
		Diagram:         NewFilePathHelper("").ConceptImageFile(conceptMap, concept),
		RelatedConcepts: conceptMap.ConceptsRelatedTo(concept),
	}
}

type SiteGenerator interface {
	GenerateSite(ctx context.Context, cmaps []*ConceptMap) error
}

// type MarkdownSiteGeneratorOption func(*MarkdownSiteGenerator)

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
		if err := sg.generateConceptMapPage(ctx, cmap); err != nil {
			return err
		}

		for _, concept := range cmap.Concepts {
			if err := sg.generateConceptPage(ctx, cmap, concept); err != nil {
				return err
			}
		}
	}

	return nil
}

func (sg *MarkdownSiteGenerator) generateConceptMapPage(ctx context.Context, cmap *ConceptMap) error {
	diagramFile := sg.filePathHelper.ConceptMapImageFile(cmap)

	if err := sg.diagramGenerator.GenerateConceptMapSVG(ctx, cmap, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMapMarkdownFile(cmap),
		conceptMapPageTemplate,
		NewConceptMapPageTemplateData(cmap))
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

func (h *FilePathHelper) ConceptMapImageFile(conceptMap *ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		"images",
		fmt.Sprintf("%s.svg", conceptMap.Slug()))
}

func (h *FilePathHelper) IndexMarkdownFile() string {
	return filepath.Join(h.BaseDir, "index.md")
}

func (h *FilePathHelper) ConceptMapMarkdownFile(conceptMap *ConceptMap) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s.md", conceptMap.Slug()))
}

func (h *FilePathHelper) ConceptImageFile(conceptMap *ConceptMap, concept *Concept) string {
	return filepath.Join(
		h.BaseDir,
		"images",
		fmt.Sprintf("%s_%s.svg", conceptMap.Slug(), concept.Key()))
}

func (h *FilePathHelper) ConceptMarkdownFile(conceptMap *ConceptMap, concept *Concept) string {
	return filepath.Join(
		h.BaseDir,
		fmt.Sprintf("%s_%s.md", conceptMap.Slug(), concept.Key()))
}
