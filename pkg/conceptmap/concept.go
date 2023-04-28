package conceptmap

import "github.com/gosimple/slug"

// Concept is a node in the concept map
type Concept struct {
	Label        string `yaml:"label"`
	Description  string `yaml:"description"`
	IsKeyConcept bool   `yaml:"isKeyConcept"`
}

// Key is normalised key of the concept
func (c *Concept) Key() string {
	return slug.Make(c.Label)
}
