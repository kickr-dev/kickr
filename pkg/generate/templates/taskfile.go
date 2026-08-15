package templates //nolint:dupl

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Taskfile returns the slice of templates related to Task configuration for a given module (build, test, docker tasks).
func Taskfile() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{"Taskfile.yml" + engine.TmplExtension},
			Out:        "Taskfile.yml",
			Remove: func(module types.Module) bool {
				if module.Dir() == types.RootModule && !module.Parent.HasBuildTool(kickr.BuildToolTask) {
					return true
				}
				return !module.HasBuildTool(kickr.BuildToolTask)
			},
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      engine.GlobsWithPart(path.Join("scripts", "task", "build.yml")),
			Out:        path.Join("scripts", "task", "build.yml"),
			Remove: func(module types.Module) bool {
				_, hugo := module.Languages[types.LanguageHugo]
				_, ok := module.Languages[types.LanguageGo]
				_, terraform := module.Languages[types.LanguageTerraform]
				return !module.HasBuildTool(kickr.BuildToolTask) || (!module.HasDocker() && !hugo && !ok && !terraform) //nolint:revive
			},
		},
	}
}

// RepositoryTaskfile returns the slice of templates related to Task configuration for the repository's root.
func RepositoryTaskfile() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join("scripts", "task", "clean.yml") + engine.TmplExtension},
			Out:        path.Join("scripts", "task", "clean.yml"),
			Remove: func(repo types.Repository) bool {
				return !repo.HasBuildTool(kickr.BuildToolTask)
			},
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join("scripts", "task", "lint.yml") + engine.TmplExtension},
			Out:        path.Join("scripts", "task", "lint.yml"),
			Remove: func(repo types.Repository) bool {
				if len(repo.ModulesWith(types.LanguageGo)) == 0 && len(repo.ModulesWith(types.LanguageTerraform)) == 0 {
					return true
				}
				return !repo.HasBuildTool(kickr.BuildToolTask)
			},
		},
	}
}
