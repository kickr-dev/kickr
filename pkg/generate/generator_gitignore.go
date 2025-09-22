package generate

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/generator"

	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GeneratorGitignore downloads and writes .gitignore file in its right path.
//
// It patches it alongside with custom kickr patches as some exclusion
// may be missing depending on kickr layout generation.
func GeneratorGitignore(httpClient *http.Client) func(ctx context.Context, destdir string, config types.KickrWrapper) error {
	if httpClient == nil {
		httpClient = http.DefaultClient //nolint:revive
	}
	return func(ctx context.Context, destdir string, config types.KickrWrapper) error {
		mapping := map[string][]string{
			"go":        {"go"},
			"helm":      {"helm"},
			"hugo":      {"hugo"},
			"node":      {"node"},
			"shell":     nil,
			"terraform": {"terraform"},
		}

		query := make([]string, 0, len(config.Languages)+3)
		for lang := range config.Languages {
			s, ok := mapping[lang]
			if ok {
				query = append(query, s...)
			}
		}
		query = append(query, "dotenv")

		if config.CI != nil {
			if slices.Contains(config.CI.Options, kickr.OptionSonarQube) {
				query = append(query, "sonar", "sonarqube")
			}
		}

		if err := generator.DownloadGitignore(ctx, httpClient, filepath.Join(destdir, generator.FileGitignore), query...); err != nil {
			return fmt.Errorf("download gitignore: %w", err)
		}

		template := engine.Template[types.KickrWrapper]{
			Delimiters: engine.DelimitersBracket(),
			Patches:    []string{".gitignore" + engine.PatchExtension + engine.TmplExtension},
			Out:        ".gitignore",
		}
		if err := engine.ApplyTemplate(templates.FS(), destdir, template, config); err != nil {
			return fmt.Errorf("apply template: %w", err)
		}
		return nil
	}
}

var _ engine.Generator[types.KickrWrapper] = GeneratorGitignore(nil) // ensure interface is implemented
