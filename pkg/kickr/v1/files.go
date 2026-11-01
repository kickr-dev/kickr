package kickr

import (
	"os"
	"path/filepath"

	"gitlab.com/kickr-dev/engine/pkg/files"

	schemas "github.com/kickr-dev/kickr/.schemas"
)

const (
	// ValuesSchema is the name of the Helm Chart values schema file.
	ValuesSchema = "values.schema.json"

	// Schema is the name of Kickr schema file to use for configurations validation.
	Schema = "kickr.v1.schema.json"
)

// Files returns the slice of available filenames for kickr configuration.
func Files() []string {
	return []string{".kickr.yml", ".kickr.yaml", ".kickr"}
}

// File returns the filename of the provided dir kickr configuration.
//
// Returns an empty string in case no kickr configuration was found inside dir.
func File(dir string) string {
	for _, file := range Files() {
		if files.Exists(filepath.Join(dir, file)) {
			return file
		}
	}
	return ""
}

// Read validates and reads the kickr configuration of dir, then ensures its defaults.
//
// Returns an empty configuration in case no kickr configuration was found inside dir.
func Read(dir string) (Kickr, error) {
	var config Kickr
	filename := File(dir)
	if filename == "" {
		return config, nil
	}

	if err := files.Validate(files.ReadYAMLFunc(schemas.FS(), Schema), files.ReadYAMLFunc(os.DirFS(dir), filename)); err != nil {
		return config, err
	}
	if err := files.ReadYAML(os.DirFS(dir), filename, &config); err != nil {
		return config, err
	}

	config.EnsureDefaults()
	return config, nil
}
