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
	var nodes []parser.PackageJSON

	// scan destdir potential node repository
	var root parser.PackageJSON
	if err := files.ReadJSON(filepath.Join(destdir, parser.FilePackageJSON), &root, os.ReadFile); err == nil {
		if err := root.Validate(); err != nil {
			return fmt.Errorf("validate '%s': %w", parser.FilePackageJSON, err)
		}
		engine.GetLogger().Infof("node detected, a '%s' is present and valid", parser.FilePackageJSON)

		repo.Module(types.RootModule).SetLanguage(types.LanguageNode, root)
		if root.Main != nil {
			repo.Module(types.RootModule).AddWorker("main") // a worker can only affected with base directory package.json
		}
		nodes = append(nodes, root)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read json: %w", err)
	}

	// scan website directory potential node repository
	if repo.Website == nil || repo.Website.Directory == "" {
		return nil
	}

	var website parser.PackageJSON
	if err := files.ReadJSON(filepath.Join(destdir, repo.Website.Directory, parser.FilePackageJSON), &website, os.ReadFile); err == nil {
		if err := website.Validate(); err != nil {
			return fmt.Errorf("validate '%s': %w", filepath.Join(repo.Website.Directory, parser.FilePackageJSON), err)
		}
		if !website.Private {
			return ErrWebsiteNoPublish
		}
		engine.GetLogger().Infof("node detected in '%s', a '%s' is present and valid", repo.Website.Directory, parser.FilePackageJSON)

		repo.Module(repo.Website.Directory).SetLanguage(types.LanguageNode, website)
		nodes = append(nodes, website)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read json in '%s': %w", repo.Website.Directory, err)
	}

	// ensure basic rules for node monorepositories are respected
	return errors.Join(HasMultipleManagers(nodes), HasMultipleRegistries(nodes))
}

var _ engine.Parser[types.Repository] = ParserNode // ensure interface is implemented
