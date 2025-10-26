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

				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{Website: &kickr.Website{Directory: dir}},
					},
				}

				// Act
				err := generate.ParserNode(ctx, destdir, &config)

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

				config := types.Repository{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{Website: &kickr.Website{Directory: dir}},
					},
				}

				// Act
				err := generate.ParserNode(ctx, destdir, &config)

				// Assert
				assert.ErrorIs(t, err, parser.ErrMissingPackageName)
				assert.ErrorIs(t, err, parser.ErrInvalidPackageManager)
			})
		}
	})

	t.Run("success_no_nodejs", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		expected := types.Repository{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
		}
		config := types.Repository{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_no_main", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6" }`), files.RwRR))

		expected := types.Repository{
			Languages: map[string]any{
				"node": generate.MonoNodes{
					{
						Directory: ".",
						Specifics: parser.PackageJSON{
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
				},
			},
		}
		config := types.Repository{}

		// Act
		err := generate.ParserNode(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_main", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`), files.RwRR))

		expected := types.Repository{
			Executables: parser.Executables{
				Workers: map[string]any{"main": struct{}{}},
			},
			Languages: map[string]any{
				"node": generate.MonoNodes{
					{
						Directory: ".",
						Specifics: parser.PackageJSON{
							Main:           func() *string { v := "index.js"; return &v }(),
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
				},
			},
		}
		config := types.Repository{}

		// Act
		err := generate.ParserNode(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
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
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6" }`), files.RwRR))

		expected := types.Repository{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
			Executables: parser.Executables{
				Workers: map[string]any{"main": struct{}{}},
			},
			Languages: map[string]any{
				"node": generate.MonoNodes{
					{
						Directory: ".",
						Specifics: parser.PackageJSON{
							Main:           func() *string { v := "index.js"; return &v }(),
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
					{
						Directory: "docs",
						Specifics: parser.PackageJSON{
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
				},
			},
		}
		config := types.Repository{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})

	t.Run("success_sub_directory", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(
			filepath.Join(destdir, "docs", parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`), files.RwRR))

		expected := types.Repository{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
			Languages: map[string]any{
				"node": generate.MonoNodes{
					{
						Directory: "docs",
						Specifics: parser.PackageJSON{
							Main:           func() *string { v := "index.js"; return &v }(),
							Name:           "kickr",
							PackageManager: "bun@1.1.6",
						},
					},
				},
			},
		}
		config := types.Repository{
			Kickr: kickr.Kickr{
				CI: &kickr.CI{Website: &kickr.Website{Directory: "docs"}},
			},
		}

		// Act
		err := generate.ParserNode(ctx, destdir, &config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, config)
	})
}
