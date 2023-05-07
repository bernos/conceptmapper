package sitegenerator

type SiteGeneratorOption func(*MarkdownSiteGenerator)

func WithDiagramGenerator(dg DiagramGenerator) SiteGeneratorOption {
	return func(sg *MarkdownSiteGenerator) {
		sg.diagramGenerator = dg
	}
}
