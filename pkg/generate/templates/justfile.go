package templates //nolint:dupl

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Justfile returns the slice of templates related to just configuration for a given module (build, test, docker recipes).
func Justfile() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{"Justfile" + engine.TmplExtension},
			Out:        "Justfile",
			Remove: func(module types.Module) bool {
				if module.Dir() == types.RootModule && !module.Parent.HasBuildTool(kickr.BuildToolJust) {
					return true
				}
				return !module.HasBuildTool(kickr.BuildToolJust)
			},
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      engine.GlobsWithPart(path.Join("scripts", "just", "build.just")),
			Out:        path.Join("scripts", "just", "build.just"),
			Remove: func(module types.Module) bool {
				_, hugo := module.Languages[types.LanguageHugo]
				_, ok := module.Languages[types.LanguageGo]
				_, terraform := module.Languages[types.LanguageTerraform]
				return !module.HasBuildTool(kickr.BuildToolJust) || (!module.HasDocker() && !hugo && !ok && !terraform) //nolint:revive
			},
		},
	}
}

// RepositoryJustfile returns the slice of templates related to just configuration for the repository's root.
func RepositoryJustfile() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join("scripts", "just", "clean.just") + engine.TmplExtension},
			Out:        path.Join("scripts", "just", "clean.just"),
			Remove: func(repo types.Repository) bool {
				return !repo.HasBuildTool(kickr.BuildToolJust)
			},
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join("scripts", "just", "lint.just") + engine.TmplExtension},
			Out:        path.Join("scripts", "just", "lint.just"),
			Remove: func(repo types.Repository) bool {
				if len(repo.ModulesWith(types.LanguageGo)) == 0 && len(repo.ModulesWith(types.LanguageTerraform)) == 0 {
					return true
				}
				return !repo.HasBuildTool(kickr.BuildToolJust)
			},
		},
	}
}
