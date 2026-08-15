package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Makefile returns the slice of templates related to make configuration for a given module (build, test, docker make tasks).
func Makefile() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{"Makefile" + engine.TmplExtension},
			Out:        "Makefile",
			Remove: func(module types.Module) bool {
				if module.Dir() == types.RootModule && !module.Parent.HasMakefile() {
					return true
				}
				return !module.HasMakefile()
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
				return !module.HasMakefile() || (!module.HasDocker() && !hugo && !ok && !terraform) //nolint:revive
			},
		},
	}
}

// RepositoryMakefile returns the slice of templates related to make configuration for the repository's root:
//
//   - build, test, docker make tasks
//   - modules Makefiles aggregation
func RepositoryMakefile() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join("scripts", "mk", "modules.mk") + engine.TmplExtension},
			Out:        path.Join("scripts", "mk", "modules.mk"),
			Remove: func(repo types.Repository) bool {
				for _, module := range repo.Modules {
					if module.Dir() != types.RootModule && module.HasMakefile() {
						return false
					}
				}
				return true
			},
		},
		{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{path.Join("scripts", "mk", "clean.mk") + engine.TmplExtension},
			Out:        path.Join("scripts", "mk", "clean.mk"),
			Remove: func(repo types.Repository) bool {
				return !repo.HasMakefile()
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
				return !repo.HasMakefile()
			},
		},
	}
}
