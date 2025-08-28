package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Golang returns the slice of templates related to Golang generation (golangci-lint, goreleaser, etc.).
func Golang() []engine.Template[types.KickrGen] {
	// Go wasn't parsed during parsers processing
	noGo := func(config types.KickrGen) bool {
		_, ok := config.Languages["go"]
		return !ok
	}

	return []engine.Template[types.KickrGen]{
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{".golangci.yml" + engine.TmplExtension},
			Out:        ".golangci.yml",
			Remove:     noGo,
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{".goreleaser.yml" + engine.TmplExtension},
			Out:        ".goreleaser.yml",
			Remove: func(config types.KickrGen) bool {
				return slices.Contains(config.Exclude, kickr.ExcludeGoreleaser) || noGo(config) || len(config.Clis) == 0 //nolint:revive
			},
		},
		{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{path.Join("internal", "build", "build.go"+engine.TmplExtension)},
			Out:        path.Join("internal", "build", "build.go"),
			Remove:     func(config types.KickrGen) bool { return noGo(config) || config.Binaries() == 0 },
		},
	}
}
