package scaffold

import (
	"embed"
	"io/fs"
	"path/filepath"
)

//go:embed aiguide/*.md aiguide/*/*.md
var aiGuideFS embed.FS

func getAIGuideFiles() (map[string]string, error) {
	files := make(map[string]string)

	err := fs.WalkDir(aiGuideFS, "aiguide", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, err := aiGuideFS.ReadFile(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("aiguide", path)
		if err != nil {
			return err
		}

		files[relPath] = string(content)
		return nil
	})

	return files, err
}

func generateAIGuide(basePath string) error {
	files, err := getAIGuideFiles()
	if err != nil {
		return err
	}

	for relPath, content := range files {
		filePath := filepath.Join(basePath, "docs/ai-guide", relPath)
		if err := writeFile(filePath, content); err != nil {
			return err
		}
	}

	return nil
}
