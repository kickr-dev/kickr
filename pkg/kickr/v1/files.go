package kickr

import (
	"path/filepath"

	"github.com/kickr-dev/engine/pkg/files"
)

const (
	// CustomValues is the name of Kickr custom Helm Chart values.
	CustomValues = "values.kickr.yaml"

	// Schema is the name of Kickr schema file to use for configurations validation.
	Schema = "kickr.v1.schema.json"
)

// Files returns the slice of available filenames for kickr configuration.
func Files() []string {
	return []string{".kickr.yml", ".kickr.yaml", ".kickr"}
}

// File returns the filepath of the provided dir kickr configuration.
//
// Returns an empty string in case no kickr configuration was found inside dir.
func File(dir string) string {
	for _, file := range Files() {
		if p := filepath.Join(dir, file); files.Exists(p) {
			return p
		}
	}
	return ""
}
