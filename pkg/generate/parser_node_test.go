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
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestParserNode(t *testing.T) {
	ctx := t.Context()

	t.Run("error_read_packagejson", func(t *testing.T) {
		for _, dir := range []string{"", "docs"} {
			t.Run(dir, func(t *testing.T) {
				// Arrange
				destdir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(destdir, dir, parser.FilePackageJSON), files.RwxRxRxRx))

				repo := types.Repository{
					Kickr: kickr.Kickr{
						Website: &kickr.Website{Directory: dir},
					},
				}

				// Act
				err := generate.ParserNode(ctx, destdir, &repo)

				// Assert
				assert.ErrorContains(t, err, "read json")
			})
		}
	})

	t.Run("error_validate_packagejson", func(t *testing.T) {
		for _, dir := range []string{"", "docs"} {
			t.Run(dir, func(t *testing.T) {
				// Arrange
				destdir := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(destdir, dir), files.RwxRxRxRx))
				require.NoError(t, os.WriteFile(filepath.Join(destdir, dir, parser.FilePackageJSON), []byte("{}"), files.RwRR))

				repo := types.Repository{
					Kickr: kickr.Kickr{
						Website: &kickr.Website{Directory: dir},
					},
				}

				// Act
				err := generate.ParserNode(ctx, destdir, &repo)

				// Assert
				assert.ErrorIs(t, err, parser.ErrMissingPackageName)
				assert.ErrorIs(t, err, parser.ErrInvalidPackageManager)
			})
		}
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
			Kickr: kickr.Kickr{
				Website: &kickr.Website{Directory: "docs"},
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
			Kickr: kickr.Kickr{
				Website: &kickr.Website{Directory: "docs"},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		assert.ErrorIs(t, err, generate.ErrWebsiteNoPublish)
	})

	t.Run("success_no_nodejs", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		expected := types.Repository{
			Kickr: kickr.Kickr{
				Website: &kickr.Website{Directory: "docs"},
			},
		}
		repo := types.Repository{
			Kickr: kickr.Kickr{
				Website: &kickr.Website{Directory: "docs"},
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

		expected := types.Repository{}
		expected.Module(types.RootModule).SetLanguage(types.LanguageNode, parser.PackageJSON{
			Name:           "kickr",
			PackageManager: "bun@1.1.6",
		})
		repo := types.Repository{}

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

		expected := types.Repository{}
		module := expected.Module(types.RootModule)
		module.Executables = parser.Executables{
			Workers: map[string]any{"main": struct{}{}},
		}
		module.SetLanguage(types.LanguageNode, parser.PackageJSON{
			Main:           func() *string { v := "index.js"; return &v }(),
			Name:           "kickr",
			PackageManager: "bun@1.1.6",
		})
		repo := types.Repository{}

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

		conf := kickr.Kickr{Website: &kickr.Website{Directory: "docs"}}
		expected := types.Repository{Kickr: conf}
		root := expected.Module(types.RootModule)
		root.Executables = parser.Executables{
			Workers: map[string]any{"main": struct{}{}},
		}
		root.SetLanguage(types.LanguageNode, parser.PackageJSON{
			Main:           func() *string { v := "index.js"; return &v }(),
			Name:           "kickr",
			PackageManager: "bun@1.1.6",
		})
		expected.Module("docs").SetLanguage(types.LanguageNode, parser.PackageJSON{
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
		})
		repo := types.Repository{Kickr: conf}

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

		conf := kickr.Kickr{Website: &kickr.Website{Directory: "docs"}}
		expected := types.Repository{Kickr: conf}
		expected.Module("docs").SetLanguage(types.LanguageNode, parser.PackageJSON{
			Main:           func() *string { v := "index.js"; return &v }(),
			Name:           "kickr",
			PackageManager: "bun@1.1.6",
			Private:        true,
		})
		repo := types.Repository{Kickr: conf}

		// Act
		err := generate.ParserNode(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}
