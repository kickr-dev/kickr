package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Golang returns the slice of templates related to Golang generation for given a module.
func Golang() []engine.Template[types.Module] {
	return []engine.Template[types.Module]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join("internal", "build", "build.go"+engine.TmplExtension)},
			Out:        path.Join("internal", "build", "build.go"),
			Remove: func(module types.Module) bool {
				_, ok := module.Languages[types.LanguageGo]
				return !ok || module.Binaries() == 0
			},
		},
	}
}

// RepositoryGolang returns the slice of templates related to Golang generation for the repository's root (golangci-lint, goreleaser, etc.).
func RepositoryGolang() []engine.Template[types.Repository] {
	return []engine.Template[types.Repository]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{".golangci.yml" + engine.TmplExtension},
			Out:        ".golangci.yml",
			Remove: func(repo types.Repository) bool {
				return !repo.HasLanguage(types.LanguageGo)
			},
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{".goreleaser.yml" + engine.TmplExtension},
			Out:        ".goreleaser.yml",
			Remove: func(repo types.Repository) bool {
				for _, module := range repo.ModulesWith(types.LanguageGo) {
					if len(module.Clis) > 0 {
						return false
					}
				}
				return true
			},
		},
	}
}
