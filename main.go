package main

import (
	"context"
	"path/filepath"

	"github.com/bernos/conceptmapping/pkg/conceptmap"
)

var (
	InputDir  = "input"
	OutputDir = "output/concept-map/docs"
)

func main() {
	ctx := context.Background()
	yamlFile := filepath.Join(InputDir, "maps.yaml")

	maps, err := conceptmap.LoadFromYamlFile(yamlFile)
	if err != nil {
		panic(err)
	}

	diagramGenerator := conceptmap.NewD2Diagram()
	siteGenererator := conceptmap.NewMarkdownSiteGenerator(diagramGenerator, OutputDir)

	if err := siteGenererator.GenerateSite(ctx, maps); err != nil {
		panic(err)
	}
}
