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
}
