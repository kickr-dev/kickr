package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/generator"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GeneratorCodeOfConduct fetches and writes the repository CODE_OF_CONDUCT.md file
// using the [Contributor Covenant 3.0] code of conduct.
//
// Placeholders notes (**NOTE: ...**) within the raw code of conduct are replaced or removed based on kickr configuration.
//
// [Contributor Covenant 3.0]: https://www.contributor-covenant.org/version/3/0/code_of_conduct
func GeneratorCodeOfConduct(httpClient *http.Client) func(ctx context.Context, destdir string, repo types.Repository) error {
	if httpClient == nil {
		httpClient = http.DefaultClient //nolint:revive
	}

	type data struct {
		Maintainers []*kickr.Maintainer
		VCS         parser.VCS
		Downloaded  string
	}

	template := engine.Template[data]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{generator.FileCodeOfConduct + engine.TmplExtension},
		Out:        generator.FileCodeOfConduct,
	}

	return func(ctx context.Context, destdir string, repo types.Repository) error {
		dest := filepath.Join(destdir, generator.FileCodeOfConduct)
		if slices.Contains(repo.Config.Exclude, kickr.ExcludeCodeOfConduct) {
			engine.GetLogger().Infof("skipping code of conduct generation, '%s' is excluded", kickr.ExcludeCodeOfConduct)
			if err := os.Remove(dest); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("remove '%s': %w", generator.FileCodeOfConduct, err)
			}
			return nil
		}

		if !engine.Forced() && files.Exists(dest) {
			engine.GetLogger().Infof("not generating '%s' since it already exists", generator.FileCodeOfConduct)
			return nil
		}

		body, err := generator.FetchCodeOfConduct(ctx, httpClient)
		if err != nil {
			return fmt.Errorf("fetch code of conduct: %w", err)
		}

		d := data{
			Downloaded:  string(body),
			Maintainers: repo.Config.Maintainers,
			VCS:         repo.VCS,
		}
		if err := engine.ApplyTemplate(templates.FS(), destdir, template, d); err != nil {
			return fmt.Errorf("apply template: %w", err)
		}
		return nil
	}
}
