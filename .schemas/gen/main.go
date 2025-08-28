package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/kickr-dev/engine/pkg/files"
)

func run() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get wd: %w", err)
	}

	paths := []string{
		filepath.Join(dir, "chart.schema.yml"),
		filepath.Join(dir, "kickr.v1.schema.yml"),
	}

	errs := make([]error, 0, 2)
	for _, file := range paths {
		in, err := os.ReadFile(file)
		if err != nil {
			errs = append(errs, fmt.Errorf("read file: %w", err))
			continue
		}

		out, err := yaml.YAMLToJSON(in)
		if err != nil {
			errs = append(errs, fmt.Errorf("yaml to json: %w", err))
			continue
		}

		if err := os.WriteFile(strings.TrimSuffix(file, filepath.Ext(file))+".json", out, files.RwRR); err != nil {
			errs = append(errs, fmt.Errorf("write file: %w", err))
			continue
		}
	}
	return errors.Join(errs...)
}

func main() {
	if err := run(); err != nil {
		slog.Error("failed to generate JSON schemas", slog.Any("error", err))
		os.Exit(1)
	}
}
