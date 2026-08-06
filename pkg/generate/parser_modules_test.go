package generate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestParserModules(t *testing.T) {
	ctx := t.Context()

	t.Run("success_implicit_root", func(t *testing.T) {
		// Arrange
		expected := types.Repository{
			Config: kickr.Kickr{Modules: []kickr.Module{{Path: "docs"}}},
			Modules: []types.Module{
				{Directory: types.RootModule, Config: kickr.Module{Path: types.RootModule}},
				{Directory: "docs", Config: kickr.Module{Path: "docs"}},
			},
		}
		for i := range expected.Modules {
			expected.Modules[i].Parent = &expected
		}
		repo := types.Repository{
			Config: kickr.Kickr{Modules: []kickr.Module{{Path: "docs"}}},
		}

		// Act
		err := generate.ParserModules(ctx, t.TempDir(), &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_explicit_modules", func(t *testing.T) {
		// Arrange
		expected := types.Repository{
			Config: kickr.Kickr{Modules: []kickr.Module{{Path: "."}, {Path: "docs"}, {Path: "apps/api"}}},
			Modules: []types.Module{
				{Directory: ".", Config: kickr.Module{Path: "."}},
				{Directory: "docs", Config: kickr.Module{Path: "docs"}},
				{Directory: "apps/api", Config: kickr.Module{Path: "apps/api"}},
			},
		}
		for i := range expected.Modules {
			expected.Modules[i].Parent = &expected
		}
		repo := types.Repository{
			Config: kickr.Kickr{Modules: []kickr.Module{{Path: "."}, {Path: "docs"}, {Path: "apps/api"}}},
		}

		// Act
		err := generate.ParserModules(ctx, t.TempDir(), &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("error_duplicate_module", func(t *testing.T) {
		// Arrange
		repo := types.Repository{
			Config: kickr.Kickr{
				Modules: []kickr.Module{
					{Path: "docs"},
					{Path: "docs/"},
				},
			},
		}

		// Act
		err := generate.ParserModules(ctx, t.TempDir(), &repo)

		// Assert
		assert.ErrorIs(t, err, generate.ErrDuplicateModule)
	})
}
