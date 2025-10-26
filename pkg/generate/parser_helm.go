package generate

import (
	"context"
	"fmt"
	"path/filepath"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// ParserHelm parses the helm chart
// and sets helm language in config by merging the config
// and .kickr overrides in chart fodler.
//
// Note, since it does marshal input configuration in JSON
// and merges it with <destdir>/chart/.kickr, this parser should be the last one called
// to ensure the configuration is in a final state.
func ParserHelm(_ context.Context, destdir string, config *types.Repository) error {
	if config.CI == nil || config.CI.Helm == nil {
		return nil
	}
	engine.GetLogger().Infof("deployment with helm detected, configuration has 'helm' key in 'deployment' section")

	base := map[string]any{
		"description": config.Description,
		"docker": func() kickr.Docker {
			if config.CI.Docker != nil {
				return *config.CI.Docker
			}
			return kickr.Docker{}
		}(),

		"clis":    config.Clis,
		"crons":   config.Crons,
		"jobs":    config.Jobs,
		"workers": config.Workers,

		"maintainers": config.Maintainers,
		"projectName": config.VCS.ProjectName,
		"projectPath": config.VCS.ProjectPath,
	}
	values, err := parser.MergeValues(base, filepath.Join(destdir, "chart", kickr.File))
	if err != nil {
		return fmt.Errorf("merge values: %w", err)
	}
	config.SetLanguage("helm", values)

	return nil
}

var _ engine.Parser[types.Repository] = ParserHelm // ensure interface is implemented
