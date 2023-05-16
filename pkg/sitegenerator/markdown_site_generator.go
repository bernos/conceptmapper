package sitegenerator

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
	"github.com/bernos/conceptmapper/pkg/diagrams"
)

type PageTemplate interface {
	Render(io.Writer) error
}

type PageTemplateFunc func(io.Writer) error

func (fn PageTemplateFunc) Render(w io.Writer) error {
	return fn(w)
}

type MarkdownSiteGenerator struct {
	diagramGenerator DiagramGenerator
	filePathHelper   *FilePathHelper
}

func NewMarkdownSiteGenerator(outputDir string, opts ...SiteGeneratorOption) *MarkdownSiteGenerator {

	sg := &MarkdownSiteGenerator{
		diagramGenerator: diagrams.NewD2DiagramGenerator(),
		filePathHelper:   NewFilePathHelper(outputDir),
	}

	for _, o := range opts {
		o(sg)
	}

	return sg
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
		NewConceptMapSummaryPageTemplate(cmap, NewFilePathHelper("../../")))
}

func (sg *MarkdownSiteGenerator) generateConceptMapDetailPage(ctx context.Context, cmap *conceptmap.ConceptMap) error {
	diagramFile := sg.filePathHelper.ConceptMapDetailImageFile(cmap)

	if err := sg.diagramGenerator.GenerateConceptMapDetailSVG(ctx, cmap, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMapDetailMarkdownFile(cmap),
		NewConceptMapDetailPageTemplate(cmap, NewFilePathHelper("../../")))
}

func (sg *MarkdownSiteGenerator) generateConceptPage(ctx context.Context, cmap *conceptmap.ConceptMap, concept *conceptmap.Concept) error {
	diagramFile := sg.filePathHelper.ConceptImageFile(cmap, concept)

	if err := sg.diagramGenerator.GenerateSingleConceptSVG(ctx, cmap, concept, diagramFile); err != nil {
		return err
	}

	return sg.renderTemplateToFile(
		sg.filePathHelper.ConceptMarkdownFile(cmap, concept),
		NewConceptPageTemplate(cmap, concept, NewFilePathHelper("../../../")))
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
