package generate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// MonoNodes is an alias of []types.Mono[parser.PackageJSON] adding various helper methods
// around the functionality of monorepositories for node.
type MonoNodes []types.Mono[parser.PackageJSON]

// HasMain returns truthy in case at least one node repository in the monorepository has a 'main' property.
//
// This helps knowing whether a node build job shall be generated or not.
func (nodes MonoNodes) HasMain() bool {
	for _, node := range nodes {
		if node.Specifics.Main != nil {
			return true
		}
	}
	return false
}

// HasMultipleManagers checks whether multiple package managers exist in the node monorepository.
func (nodes MonoNodes) HasMultipleManagers() bool {
	managers := make(map[string]struct{}, len(nodes))
	for _, mono := range nodes {
		manager, _, _ := strings.Cut(mono.Specifics.PackageManager, "@")
		managers[manager] = struct{}{}
	}
	return len(managers) > 1
}

// ErrMultipleManagers is returned when there's a monorepo involved for node language
// and multiple package managers are defined between parsed repositories.
//
// This error exists to ensure consistency for package managers inside one git repository, which shouldn't be that hard to aim.
var ErrMultipleManagers = errors.New("multiple node package manager")

// ParserNode detects the presence of a ParserNode.js project by looking for a package.json file.
//
// In case of success, the function will set the language to "node"
// and the worker to "main" if the main property is present in the package.json file.
func ParserNode(_ context.Context, destdir string, config *types.KickrWrapper) error {
	monos := make(MonoNodes, 0, 2)

	// scan destdir potential node repository
	var root parser.PackageJSON
	err := files.ReadJSON(filepath.Join(destdir, parser.FilePackageJSON), &root, os.ReadFile)
	if err == nil {
		if err := root.Validate(); err != nil {
			return fmt.Errorf("validate '%s': %w", parser.FilePackageJSON, err)
		}
		engine.GetLogger().Infof("node detected, a '%s' is present and valid", parser.FilePackageJSON)

		monos = append(monos, types.Mono[parser.PackageJSON]{Directory: ".", Specifics: root})
		if root.Main != nil {
			config.AddWorker("main") // a worker can only affected with base directory package.json
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read json: %w", err)
	}

	// scan website directory potential node repository
	if config.CI != nil && config.CI.Website != nil && config.CI.Website.Directory != "" {
		var website parser.PackageJSON
		err := files.ReadJSON(filepath.Join(destdir, config.CI.Website.Directory, parser.FilePackageJSON), &website, os.ReadFile)
		if err == nil {
			if err := website.Validate(); err != nil {
				return fmt.Errorf("validate '%s': %w", parser.FilePackageJSON, err)
			}
			engine.GetLogger().Infof("node detected in '%s', a '%s' is present and valid", config.CI.Website.Directory, parser.FilePackageJSON)

			monos = append(monos, types.Mono[parser.PackageJSON]{Directory: config.CI.Website.Directory, Specifics: website})
		} else if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("read json in '%s': %w", config.CI.Website.Directory, err)
		}
	}

	// avoid handling multiple package managers with monorepository setup
	if monos.HasMultipleManagers() {
		return ErrMultipleManagers
	}
	if len(monos) > 0 {
		config.SetLanguage("node", monos)
	}
	return nil
}

var _ engine.Parser[types.KickrWrapper] = ParserNode // ensure interface is implemented
