package generate

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// ErrDuplicateModule is returned when two 'modules' entries resolve to the same path.
var ErrDuplicateModule = errors.New("duplicate module path")

// ParserModules seeds the repository modules from the 'modules' kickr configuration.
//
// When 'modules' is absent or empty, the repository root is the sole module.
func ParserModules(_ context.Context, _ string, repo *types.Repository) error {
	seen := make(map[string]struct{}, len(repo.Config.Modules))
	errs := make([]error, 0, len(repo.Config.Modules))
	for _, module := range repo.Config.Modules {
		path := filepath.Clean(module.Path)
		if _, ok := seen[path]; ok {
			errs = append(errs, fmt.Errorf("%w '%s'", ErrDuplicateModule, path))
			continue
		}
		seen[path] = struct{}{}

		repo.Modules = append(repo.Modules, types.Module{
			Config:    module,
			Directory: path,
			Parent:    repo,
		})
	}

	// root is a mandatory module and always the first one
	if i := repo.ModuleIndexOf(types.RootModule); i < 0 {
		repo.Modules = slices.Insert(repo.Modules, 0, types.Module{
			Config:    kickr.Module{Path: types.RootModule},
			Directory: types.RootModule,
			Parent:    repo,
		})
	}

	return errors.Join(errs...) // already wrapped
}

var _ engine.Parser[types.Repository] = ParserModules // ensure interface is implemented
