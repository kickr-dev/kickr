package generate_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
)

func TestParserGolang(t *testing.T) {
	ctx := t.Context()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

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
		err = generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_image", func(t *testing.T) {
		// Arrange
		t.Setenv("DOCKER_CONFIG", t.TempDir())

		manifest := "https://registry-1.docker.io/v2/library/golang/manifests/1.25.3-trixie"
		digest := "sha256:71d7c66aa2305f7d239c35d3c04b688c709ee0243d792afb868a3731494f08bb"
		httpmock.RegisterResponder(http.MethodHead, manifest,
			httpmock.NewStringResponder(http.StatusOK, "{}").
				HeaderSet(http.Header{"Content-Type": {"application/vnd.oci.image.index.v1+json"}, "Docker-Content-Digest": {digest}}).
				SetContentLength())

		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte(
			`go 1.22

			toolchain go1.25.3

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
							ContainerImage: generate.GoImage + ":1.25.3-trixie@" + digest,
							Go:             "1.22",
							Toolchain:      "1.25.3",
							Uses:           []parser.GoworkUse{{Gomod: libmod, Use: "./lib1"}},
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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})

	t.Run("success_image_toolchain_default", func(t *testing.T) {
		// Arrange
		t.Setenv("DOCKER_CONFIG", t.TempDir())

		manifest := "https://registry-1.docker.io/v2/library/golang/manifests/1.22-trixie"
		digest := "sha256:71d7c66aa2305f7d239c35d3c04b688c709ee0243d792afb868a3731494f08bb"
		httpmock.RegisterResponder(http.MethodHead, manifest,
			httpmock.NewStringResponder(http.StatusOK, "{}").
				HeaderSet(http.Header{"Content-Type": {"application/vnd.oci.image.index.v1+json"}, "Docker-Content-Digest": {digest}}).
				SetContentLength())

		destdir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(destdir, parser.FileGomod), []byte(
			`module github.com/kickr-dev/kickr

			go 1.22

			toolchain default`,
		), files.RwRR))

		expected := types.Repository{
			Modules: []types.Module{
				{
					Directory: types.RootModule,
					Languages: map[string]any{
						types.LanguageGo: parser.Gomod{
							ContainerImage: generate.GoImage + ":1.22-trixie@" + digest,
							Go:             "1.22",
							Module:         "github.com/kickr-dev/kickr",
							Toolchain:      "default",
							Tools:          []string{},
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
		err := generate.ParserGolang(http.DefaultClient)(ctx, destdir, &repo)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, repo)
	})
}
