package types_test

import (
	"testing"

	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestModule(t *testing.T) {
	t.Run("success_sorted", func(t *testing.T) {
		// Arrange
		repo := types.Repository{
			Kickr: kickr.Kickr{Version: 1},
		}

		// Act
		repo.Module("docs")
		repo.Module(types.RootModule)
		repo.Module("apps/api")

		// Assert
		dirs := []string{types.RootModule, "apps/api", "docs"}
		for _, module := range repo.Modules() {
			assert.Same(t, &repo, module.Repository())
			assert.Equal(t, dirs[0], module.Dir())
			dirs = dirs[1:]
		}
	})

	t.Run("success_same_directory_returns_same_module", func(t *testing.T) {
		// Arrange
		repo := types.Repository{}
		gomod := parser.Gomod{Module: "github.com/kickr-dev/kickr"}

		// Act
		repo.Module(types.RootModule).SetLanguage(types.LanguageGo, gomod)

		// Assert
		assert.Len(t, repo.Modules(), 1)
	})

	t.Run("success_is_website", func(t *testing.T) {
		// Arrange
		repo := types.Repository{Kickr: kickr.Kickr{Website: &kickr.Website{Directory: "docs"}}}

		// Act
		docs := repo.Module("docs")
		root := repo.Module(types.RootModule)

		// Assert
		assert.True(t, docs.IsWebsite())
		assert.False(t, root.IsWebsite())
	})

	t.Run("success_is_website_no_website_configured", func(t *testing.T) {
		// Arrange
		repo := types.Repository{}

		// Act
		root := repo.Module(types.RootModule)

		// Assert
		assert.False(t, root.IsWebsite())
	})

	t.Run("success_has_docker_go_with_binaries", func(t *testing.T) {
		// Arrange
		repo := types.Repository{Kickr: kickr.Kickr{
			Docker:  &kickr.Docker{Exclude: []string{kickr.DockerExcludeWebsite}},
			Website: &kickr.Website{Directory: types.RootModule},
		}}

		// Act
		module := repo.Module(types.RootModule).SetLanguage(types.LanguageGo, nil)
		module.AddCLI("cli")

		// Assert
		assert.True(t, module.HasDocker())
	})

	t.Run("success_has_docker_go_without_binaries", func(t *testing.T) {
		// Arrange
		repo := types.Repository{Kickr: kickr.Kickr{Docker: &kickr.Docker{}}}

		// Act
		module := repo.Module(types.RootModule).SetLanguage(types.LanguageGo, nil)

		// Assert
		assert.False(t, module.HasDocker())
	})

	t.Run("success_has_docker_website_excluded", func(t *testing.T) {
		// Arrange
		repo := types.Repository{Kickr: kickr.Kickr{
			Docker:  &kickr.Docker{Exclude: []string{kickr.DockerExcludeWebsite}},
			Website: &kickr.Website{Directory: "docs"},
		}}

		// Act
		module := repo.Module("docs").SetLanguage(types.LanguageHugo, nil)

		// Assert
		assert.False(t, module.HasDocker())
	})

	t.Run("success_has_docker_website_not_excluded", func(t *testing.T) {
		// Arrange
		repo := types.Repository{Kickr: kickr.Kickr{
			Docker:  &kickr.Docker{},
			Website: &kickr.Website{Directory: "docs"},
		}}

		// Act
		module := repo.Module("docs").SetLanguage(types.LanguageNode, nil)

		// Assert
		assert.True(t, module.HasDocker())
	})

	t.Run("success_has_docker_non_website_module", func(t *testing.T) {
		// Arrange
		repo := types.Repository{Kickr: kickr.Kickr{
			Docker:  &kickr.Docker{Exclude: []string{kickr.DockerExcludeWebsite}},
			Website: &kickr.Website{Directory: "docs"},
		}}

		// Act
		module := repo.Module(types.RootModule).SetLanguage(types.LanguageNode, nil)

		// Assert
		assert.True(t, module.HasDocker())
	})

	t.Run("success_has_docker_no_docker_configured", func(t *testing.T) {
		// Arrange
		repo := types.Repository{}

		// Act
		module := repo.Module(types.RootModule).SetLanguage(types.LanguageGo, nil)

		// Assert
		assert.False(t, module.HasDocker())
	})
}
