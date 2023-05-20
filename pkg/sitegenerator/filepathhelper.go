package sitegenerator

import (
	"fmt"
	"path/filepath"

	"github.com/bernos/conceptmapper/pkg/conceptmap"
)

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
