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

var (
	// ErrMultipleManagers is returned when there's a monorepo involved for node language
	// and multiple package managers are defined between parsed repositories.
	//
	// This error exists to ensure consistency for package managers inside one git repository, which shouldn't be that hard to aim.
	ErrMultipleManagers = errors.New("multiple node package managers")

	// ErrMultipleRegistries is returned when there's a monorepo involved for node language
	// and multiple registries are defined between parsed repositories.
	//
	// This error exists to ensure consistency for registries inside one git repository, which shouldn't be that hard to aim.
	ErrMultipleRegistries = errors.New("multiple node registries")

	// ErrModuleNoPublish is returned when a non-root node module has publishing enabled.
	//
	// Due to limitations regarding semantic-release configuration (.releaserc, GitLab CICD setup),
	// only the root node module can be published.
	ErrModuleNoPublish = errors.New("only the root node module can be published")
)

// HasMultipleManagers returns an error in case multiple package managers exist in the node monorepository.
func HasMultipleManagers(nodes []parser.PackageJSON) error {
	managers := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		manager, _, _ := strings.Cut(node.PackageManager, "@")
		managers[manager] = struct{}{}
	}
	if len(managers) > 1 {
		return ErrMultipleManagers
	}
	return nil
}

// HasMultipleRegistries returns an error in case multiple registries exist in the node monorepository.
func HasMultipleRegistries(nodes []parser.PackageJSON) error {
	registries := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		if node.Private {
			continue
		}
		registries[node.PublishConfig.Registry] = struct{}{}
	}
	if len(registries) > 1 {
		return ErrMultipleRegistries
	}
	return nil
}

// ParserNode detects the presence of a ParserNode.js project by looking for a package.json file.
//
// In case of success, the function will set the language to "node"
// and the worker to "main" if the main property is present in the package.json file.
func ParserNode(_ context.Context, destdir string, repo *types.Repository) error {
	nodes := make([]parser.PackageJSON, 0, len(repo.Modules))
	for i, module := range repo.Modules {
		var node parser.PackageJSON
		if err := files.ReadJSON(filepath.Join(destdir, module.Dir(), parser.FilePackageJSON), &node, os.ReadFile); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("read json in '%s': %w", module.Dir(), err)
			}
			continue
		}
		if err := node.Validate(); err != nil {
			return fmt.Errorf("validate '%s': %w", filepath.Join(module.Dir(), parser.FilePackageJSON), err)
		}
		if !node.Private && module.Dir() != types.RootModule {
			return ErrModuleNoPublish
		}
		engine.GetLogger().Infof("node detected in '%s', a '%s' is present and valid", module.Dir(), parser.FilePackageJSON)

		repo.Modules[i].SetLanguage(types.LanguageNode, node)
		if node.Main != nil {
			repo.Modules[i].AddWorker("main")
		}
		nodes = append(nodes, node)
	}

	// ensure basic rules for node monorepositories are respected
	return errors.Join(HasMultipleManagers(nodes), HasMultipleRegistries(nodes))
}

var _ engine.Parser[types.Repository] = ParserNode // ensure interface is implemented
