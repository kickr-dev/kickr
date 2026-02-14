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

	// ErrWebsiteNoPublish is returned when a website is provided through kickr configuration
	// and is a node repository but with publishing enabled.
	//
	// Due to limitations regarding semantic-release configuration (.releaserc, GitLab CICD setup),
	// only the root node repository can be published.
	ErrWebsiteNoPublish = errors.New("website node module should be private")
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

// HasMultipleManagers returns an error in case multiple package managers exist in the node monorepository.
func (nodes MonoNodes) HasMultipleManagers() error {
	managers := make(map[string]struct{}, len(nodes))
	for _, mono := range nodes {
		manager, _, _ := strings.Cut(mono.Specifics.PackageManager, "@")
		managers[manager] = struct{}{}
	}
	if len(managers) > 1 {
		return ErrMultipleManagers
	}
	return nil
}

// HasMultipleRegistries returns an error in case multiple registries exist in the node monorepository.
func (nodes MonoNodes) HasMultipleRegistries() error {
	registries := make(map[string]struct{}, len(nodes))
	for _, mono := range nodes {
		if mono.Specifics.Private {
			continue
		}
		registries[mono.Specifics.PublishConfig.Registry] = struct{}{}
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
func ParserNode(_ context.Context, destdir string, config *types.Repository) error {
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
	if config.Website != nil && config.Website.Directory != "" {
		var website parser.PackageJSON
		err := files.ReadJSON(filepath.Join(destdir, config.Website.Directory, parser.FilePackageJSON), &website, os.ReadFile)
		if err == nil {
			if err := website.Validate(); err != nil {
				return fmt.Errorf("validate '%s': %w", filepath.Join(config.Website.Directory, parser.FilePackageJSON), err)
			}
			if !website.Private {
				return ErrWebsiteNoPublish
			}
			engine.GetLogger().Infof("node detected in '%s', a '%s' is present and valid", config.Website.Directory, parser.FilePackageJSON)

			monos = append(monos, types.Mono[parser.PackageJSON]{Directory: config.Website.Directory, Specifics: website})
		} else if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("read json in '%s': %w", config.Website.Directory, err)
		}
	}

	// ensure basic rules for node monorepositories are respected
	if err := errors.Join(monos.HasMultipleManagers(), monos.HasMultipleRegistries()); err != nil {
		return err
	}

	if len(monos) > 0 {
		config.SetLanguage("node", monos)
	}
	return nil
}

var _ engine.Parser[types.Repository] = ParserNode // ensure interface is implemented
