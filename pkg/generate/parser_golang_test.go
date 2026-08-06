package generate_test

import (
	"fmt"
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

func TestParserHugo(t *testing.T) {
	ctx := t.Context()

	t.Run("error_parse_hugo", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(destdir, files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "hugo.toml"), []byte("{ invalid toml }"), files.RwRR))

		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserHugo(ctx, destdir, &repo)

		// Assert
		require.NotErrorIs(t, err, parser.ErrNoHugo)
		assert.ErrorContains(t, err, "read hugo in '.'")
	})

	t.Run("success_no_hugo", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		expected := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserHugo(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_root", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		hugoconfig, err := os.Create(filepath.Join(destdir, "hugo.toml"))
		require.NoError(t, err)
		require.NoError(t, hugoconfig.Close())

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{types.LanguageHugo: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err = generate.ParserHugo(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_docs", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx))
		hugoconfig, err := os.Create(filepath.Join(destdir, "docs", "hugo.toml"))
		require.NoError(t, err)
		require.NoError(t, hugoconfig.Close())

		expected := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{
					Directory: "docs",
					Languages: map[string]any{types.LanguageHugo: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
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
		err = generate.ParserHugo(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}

func TestParserGolang(t *testing.T) {
	ctx := t.Context()

	t.Run("error_read_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(destdir, parser.FileGomod), files.RwxRxRxRx))

		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserGolang(ctx, destdir, &repo)

		// Assert
		assert.ErrorContains(t, err, fmt.Sprintf("read '%s'", parser.FileGomod))
	})

	t.Run("skip_when_root_has_hugo", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FileGomod), []byte(
			`module github.com/kickr-dev/kickr

			go 1.22`,
		), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{types.LanguageHugo: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{types.LanguageHugo: parser.HugoCompose{HugoConfig: &parser.HugoConfig{}}},
				},
			},
		}

		// Act
		err := generate.ParserGolang(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("error_no_use_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte(
			`go 1.22

			use (
				./lib1
			)`,
		), files.RwRR))

		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserGolang(ctx, destdir, &repo)

		// Assert
		require.NotErrorIs(t, err, parser.ErrNoGowork)
		require.ErrorContains(t, err, "read 'go.work'")
	})

	t.Run("success_no_gowork_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		expected := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
			},
		}

		// Act
		err := generate.ParserGolang(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_gomod", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FileGomod), []byte(
			`module github.com/kickr-dev/kickr

			go 1.22

			tool (
				example.com/tool-example
			)`,
		), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{
						types.LanguageGo: parser.Gomod{
							Module: "github.com/kickr-dev/kickr",
							Go:     "1.22",
							Tools:  []string{"example.com/tool-example"},
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
		err := generate.ParserGolang(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_gowork", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte(
			`go 1.22

			use (
				./lib1
			)`,
		), files.RwRR))

		require.NoError(t, os.MkdirAll(filepath.Join(destdir, "lib1"), files.RwxRxRxRx))
		require.NoError(t, os.WriteFile(filepath.Join(destdir, "lib1", parser.FileGomod), []byte("module github.com/kickr-dev/kickr\ngo 1.22"), files.RwRR))

		libmod := parser.Gomod{
			Go:     "1.22",
			Module: "github.com/kickr-dev/kickr",
			Tools:  []string{},
		}
		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{
						types.LanguageGo: parser.Gowork{
							Go:   "1.22",
							Uses: []parser.GoworkUse{{Gomod: libmod, Use: "./lib1"}},
						},
					},
				},
				{
					Directory: "lib1",
					Languages: map[string]any{types.LanguageGo: libmod},
				},
			},
		}
		repo := types.Repository{
			Modules: []types.Module{
				{Directory: types.RootModule},
				{Directory: "lib1"},
			},
		}

		// Act
		err := generate.ParserGolang(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_gomod_cmd", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FileGomod), []byte(
			`module github.com/kickr-dev/kickr

			go 1.22`,
		), files.RwRR))

		cmd := filepath.Join(destdir, parser.FolderCMD)
		require.NoError(t, os.MkdirAll(cmd, files.RwxRxRxRx))
		cli := filepath.Join(cmd, "name")
		require.NoError(t, os.MkdirAll(cli, files.RwxRxRxRx))
		main, err := os.Create(filepath.Join(cli, parser.FileMain))
		require.NoError(t, err)
		require.NoError(t, main.Close())

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory:   types.RootModule,
					Executables: parser.Executables{Clis: map[string]any{"name": struct{}{}}},
					Languages: map[string]any{
						types.LanguageGo: parser.Gomod{
							Module: "github.com/kickr-dev/kickr",
							Go:     "1.22",
							Tools:  []string{},
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
		err = generate.ParserGolang(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}
