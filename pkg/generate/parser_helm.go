package generate

import (
	"context"
	"fmt"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"
	"gitlab.com/kickr-dev/workload"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// ParserHelm sets the helm language on the root module with the Helm Chart values
// (description, maintainers and the executables of every module deployed with helm).
//
// Note, since it reads modules executables, this parser should be the last one called
// to ensure the configuration is in a final state.
func ParserHelm(_ context.Context, _ string, repo *types.Repository) error {
	if repo.Config.Helm == nil {
		return nil
	}
	engine.GetLogger().Infof("deployment with helm detected, configuration has 'helm' key in 'deployment' section")

	modules := repo.ModulesWithDeployment(kickr.DeploymentTargetHelm)

	path := repo.VCS.ProjectPath
	if repo.Config.Docker != nil && repo.Config.Docker.Path != "" {
		path = repo.Config.Docker.Path
	}

	executables := func(get func(types.Module) map[string]any) map[string]any {
		result := map[string]any{}
		for _, module := range modules {
			repository := path
			if slug := module.Slug(); slug != "" {
				repository += "/" + slug
			}
			for name := range get(module) {
				result[name] = map[string]any{"image": map[string]any{"repository": repository}}
			}
		}
		return result
	}

	base := map[string]any{
		"description": repo.Config.Description,
		"docker": func() kickr.Docker {
			if repo.Config.Docker != nil {
				return *repo.Config.Docker
			}
			return kickr.Docker{}
		}(),
		"workloadVersion": workload.Version(),

		"clis":    executables(func(m types.Module) map[string]any { return m.Clis }),
		"crons":   executables(func(m types.Module) map[string]any { return m.Crons }),
		"jobs":    executables(func(m types.Module) map[string]any { return m.Jobs }),
		"workers": executables(func(m types.Module) map[string]any { return m.Workers }),

		"maintainers": repo.Config.Maintainers,
		"projectName": repo.VCS.ProjectName,
		"projectPath": repo.VCS.ProjectPath,
	}

	// Templates read values with their json names (e.g. docker.registry).
	values, err := parser.MergeValues(base)
	if err != nil {
		return fmt.Errorf("merge values: %w", err)
	}
	if i := repo.ModuleIndexOf(types.RootModule); i >= 0 {
		repo.Modules[i].SetLanguage(types.LanguageHelm, values)
	}
	return nil
}

var _ engine.Parser[types.Repository] = ParserHelm // ensure interface is implemented
