package conceptmap

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type yamlDefinition struct {
	Title        string            `yaml:"title"`
	Description  string            `yaml:"description"`
	Propositions string            `yaml:"propositions"`
	Concepts     map[string]string `yaml:"concepts"`
}

// LoadFromYamlFile loads a Map from a yaml file
func LoadFromYamlFile(file string) ([]*ConceptMap, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return LoadFromYamlReader(f)
}

// LoadFromYamlReader loads a Map from an io.Reader in yaml format
func LoadFromYamlReader(r io.Reader) ([]*ConceptMap, error) {
	dec := yaml.NewDecoder(r)
	out := []*ConceptMap{}
	parser := &PropositionParser{}

	for {
		def := new(yamlDefinition)

		err := dec.Decode(def)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, err
			}
		}

		m := &ConceptMap{
			Title:        def.Title,
			Description:  def.Description,
			Concepts:     []*Concept{},
			Propositions: []*Proposition{},
		}

		if err := parser.Parse(def.Propositions, &m.Propositions, &m.Concepts); err != nil {
			return nil, err
		}

		for k, v := range def.Concepts {
			for _, c := range m.Concepts {
				if c.Label == k {
					c.Description = v
				}
			}
		}

		out = append(out, m)
	}

	return out, nil
}
