package generate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

func TestParserNode(t *testing.T) {
	ctx := t.Context()

	t.Run("error_read_packagejson", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, parser.FilePackageJSON), files.RwxRxRxRx))
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		assert.ErrorContains(t, err, "read json")
	})

	t.Run("error_validate_packagejson", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(destdir, files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), []byte("{}"), files.RwRR))
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		assert.ErrorIs(t, err, parser.ErrMissingPackageName)
		assert.ErrorIs(t, err, parser.ErrInvalidPackageManager)
	})

	t.Run("error_consistencies", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "publishConfig": { "registry": "npm.example.com" } }`), files.RwRR))
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(filepath.Join(destdir, "docs"), parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "pnpm@1.1.6", "private": true }`), files.RwRR))

		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "docs"},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		assert.ErrorIs(t, err, generate.ErrMultipleManagers)
	})

	t.Run("error_sub_directory_published", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(filepath.Join(destdir, "docs"), parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "pnpm@1.1.6" }`), files.RwRR))

		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "docs"},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		assert.ErrorIs(t, err, generate.ErrModuleNoPublish)
	})

	t.Run("success_no_nodejs", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		expected := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "docs"},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "docs"},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_no_main", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6" }`), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{
						types.LanguageNode: parser.PackageJSON{
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_main", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory:   types.RootModule,
					Executables: parser.Executables{Workers: map[string]any{"main": struct{}{}}},
					Languages: map[string]any{
						types.LanguageNode: parser.PackageJSON{
							Main:           func() *string { v := "index.js"; return &v }(),
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_root_and_sub_directory", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`), files.RwRR))
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, "docs", parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "private": true, "publishConfig": { "registry": "npm.example.org" } }`), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory:   types.RootModule,
					Executables: parser.Executables{Workers: map[string]any{"main": struct{}{}}},
					Languages: map[string]any{
						types.LanguageNode: parser.PackageJSON{
							Main:           func() *string { v := "index.js"; return &v }(),
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
				},
				{
					Directory: "docs",
					Languages: map[string]any{
						types.LanguageNode: parser.PackageJSON{
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
							Private:        true,
							PublishConfig: struct {
								Access     string `json:"access,omitempty"`
								Provenance bool   `json:"provenance,omitempty"`
								Registry   string `json:"registry,omitempty"`
								Tag        string `json:"tag,omitempty"`
							}{
								Registry: "npm.example.org",
							},
						},
					},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "docs"},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_sub_directory", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, "docs", parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js", "private": true }`), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{
					Directory:   "docs",
					Executables: parser.Executables{Workers: map[string]any{"main": struct{}{}}},
					Languages: map[string]any{
						types.LanguageNode: parser.PackageJSON{
							Main:           func() *string { v := "index.js"; return &v }(),
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
							Private:        true,
						},
					},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "docs"},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}
