package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
)

func run() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get wd: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	errs := make([]error, 0, 2)
	for _, file := range entries {
		filename := file.Name()
		if !slices.Contains([]string{".yml", ".yaml"}, filepath.Ext(filename)) {
			continue // ignore non-ya?ml files
		}

		in, err := os.ReadFile(filename)
		if err != nil {
			errs = append(errs, fmt.Errorf("read file: %w", err))
			continue
		}

		out, err := yaml.YAMLToJSON(in)
		if err != nil {
			errs = append(errs, fmt.Errorf("yaml to json: %w", err))
			continue
		}

		if err := os.WriteFile(strings.TrimSuffix(filename, filepath.Ext(filename))+".json", out, 0o644); err != nil {
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
