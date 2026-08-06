package generate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/generator"

	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// GeneratorGitignore fetches the languages related exclusions from https://docs.gitignore.io/use/api
// and writes a .gitignore file in every module of the repository.
//
// Each module only gets the exclusions of the languages it's made of,
// so a hugo module in a subdirectory gets the hugo exclusions in its own .gitignore, not the repository root one.
//
// The fetched content is appended to the kickr specific exclusions,
// as some of them may be missing depending on kickr layout generation.
func GeneratorGitignore(httpClient *http.Client) func(ctx context.Context, destdir string, repo types.Repository) error {
	if httpClient == nil {
		httpClient = http.DefaultClient //nolint:revive
	}

	type data struct {
		types.Module //nolint:embeddedstructfieldcheck
		Downloaded   string
	}

	parameters := map[string][]string{
		types.LanguageGo:        {"go"},
		types.LanguageHelm:      {"helm"},
		types.LanguageHugo:      {"hugo"},
		types.LanguageNode:      {"node"},
		types.LanguageTerraform: {"terraform"},
	}

	return func(ctx context.Context, destdir string, repo types.Repository) error {
		template := engine.Template[data]{
			Delimiters:     engine.DelimitersBracket(),
			GeneratePolicy: engine.PolicyAlways,
			Globs:          []string{generator.FileGitignore + engine.TmplExtension},
			Out:            generator.FileGitignore,
		}

		// modules sharing the same languages share the same payload, no need to fetch it twice
		ignores := make(map[string]string, len(repo.Modules))
		errs := make([]error, 0, len(repo.Modules))
		for _, module := range repo.Modules {
			query := make([]string, 0, len(module.Languages)+3)
			for language := range module.Languages {
				query = append(query, parameters[language]...)
			}
			query = append(query, "dotenv")
			if module.Dir() == types.RootModule && repo.Config.HasSonarQube() { // sonar analyzes the whole repository from its root, wherever the analyzed code lives
				query = append(query, "sonar", "sonarqube")
			}

			slices.Sort(query) // languages come from a map, the query must stay stable across runs
			key := strings.Join(query, ":")
			if _, ok := ignores[key]; !ok {
				body, err := generator.FetchGitignore(ctx, httpClient, query...)
				if err != nil {
					errs = append(errs, fmt.Errorf("download gitignore for '%s': %w", module.Dir(), err))
					continue
				}
				ignores[key] = string(body)
			}

			if err := engine.ApplyTemplate(templates.FS(), filepath.Join(destdir, module.Dir()), template, data{Module: module, Downloaded: ignores[key]}); err != nil {
				errs = append(errs, fmt.Errorf("apply template in '%s': %w", module.Dir(), err))
			}
		}
		return errors.Join(errs...)
	}
}

var _ engine.Generator[types.Repository] = GeneratorGitignore(nil) // ensure interface is implemented
