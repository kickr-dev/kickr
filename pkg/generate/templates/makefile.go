package templates //nolint:dupl

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Makefile returns the slice of templates related to make configuration for a given module (build, test, docker make tasks).
func Makefile() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"Makefile" + engine.TmplExtension},
			Out:        "Makefile",
			Remove: func(module types.Module) bool {
				if module.Dir() == types.RootModule && !module.Parent.HasBuildTool(kickr.BuildToolMake) {
					return true
				}
				return !module.HasBuildTool(kickr.BuildToolMake)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      engine.GlobsWithPart(path.Join("scripts", "mk", "build.mk")),
			Out:        path.Join("scripts", "mk", "build.mk"),
			Remove: func(module types.Module) bool {
				_, hugo := module.Languages[types.LanguageHugo]
				_, ok := module.Languages[types.LanguageGo]
				_, terraform := module.Languages[types.LanguageTerraform]
				return !module.HasBuildTool(kickr.BuildToolMake) || (!module.HasDocker() && !hugo && !ok && !terraform) //nolint:revive
			},
		},
	}
}

// RepositoryMakefile returns the slice of templates related to make configuration for the repository's root.
func RepositoryMakefile() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join("scripts", "mk", "clean.mk") + engine.TmplExtension},
			Out:        path.Join("scripts", "mk", "clean.mk"),
			Remove: func(repo types.Repository) bool {
				return !repo.HasBuildTool(kickr.BuildToolMake)
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join("scripts", "mk", "lint.mk") + engine.TmplExtension},
			Out:        path.Join("scripts", "mk", "lint.mk"),
			Remove: func(repo types.Repository) bool {
				if len(repo.ModulesWith(types.LanguageGo)) == 0 && len(repo.ModulesWith(types.LanguageTerraform)) == 0 {
					return true
				}
				return !repo.HasBuildTool(kickr.BuildToolMake)
			},
		},
	}
}
