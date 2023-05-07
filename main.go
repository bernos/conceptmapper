package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
	"github.com/bernos/conceptmapper/pkg/diagrams"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx := context.Background()

	app := &cli.App{
		Name:  "conceptmapper",
		Usage: "Build concept maps",

		Commands: []*cli.Command{
			{
				Name: "generate-markdown-site",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "outdir",
						Aliases: []string{"o"},
						Usage:   "Output markdown site to this dir",
					},
				},
				Action: func(c *cli.Context) error {
					outputDir := c.String("outdir")
					inputFile := c.Args().Get(0)

					if outputDir == "" {
						return fmt.Errorf("outdir is required")
					}

					if inputFile == "" {
						return fmt.Errorf("input file is required")
					}

					maps, err := conceptmap.LoadFromYamlFile(inputFile)
					if err != nil {
						return err
					}

					diagramGenerator := diagrams.NewD2DiagramGenerator(
						diagrams.WithDirection(diagrams.DirectionDown))

					siteGenererator := conceptmap.NewMarkdownSiteGenerator(diagramGenerator, outputDir)

					return siteGenererator.GenerateSite(ctx, maps)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
