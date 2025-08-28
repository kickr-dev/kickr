package generate_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/generator"
	"github.com/kickr-dev/engine/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

func TestGeneratorLicense_Remove(t *testing.T) {
	ctx := t.Context()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	gen := generate.GeneratorLicense(http.DefaultClient)

	t.Run("error_remove_license", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		dest := filepath.Join(destdir, generator.FileLicense)
		require.NoError(t, os.MkdirAll(filepath.Join(dest, "file.txt"), files.RwxRxRxRx))

		// Act
		err := gen(ctx, destdir, types.KickrGen{})

		// Assert
		assert.ErrorContains(t, err, fmt.Sprintf("remove '%s'", generator.FileLicense))
	})

	t.Run("success_remove_no_license", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		dest := filepath.Join(destdir, generator.FileLicense)

		// Act
		err := gen(ctx, destdir, types.KickrGen{})

		// Assert
		require.NoError(t, err)
		assert.NoFileExists(t, dest)
	})

	t.Run("success_remove_license", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		dest := filepath.Join(destdir, generator.FileLicense)
		license, err := os.Create(dest)
		require.NoError(t, err)
		require.NoError(t, license.Close())

		// Act
		err = gen(ctx, destdir, types.KickrGen{})

		// Assert
		require.NoError(t, err)
		assert.NoFileExists(t, dest)
	})
}

func TestGeneratorLicense_Download(t *testing.T) {
	ctx := t.Context()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	gen := generate.GeneratorLicense(http.DefaultClient)

	url := generator.GitLabURL + "/templates/licenses/mit"

	t.Run("error_http_call", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		httpmock.RegisterResponder(http.MethodGet, url,
			httpmock.NewStringResponder(http.StatusInternalServerError, "error message"))

		// Act
		err := gen(ctx, t.TempDir(), types.KickrGen{Kickr: kickr.Kickr{License: "mit"}})

		// Assert
		assert.ErrorContains(t, err, "download license")
	})

	t.Run("success_already_exists", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		dest := filepath.Join(destdir, generator.FileLicense)
		license, err := os.Create(dest)
		require.NoError(t, err)
		require.NoError(t, license.Close())

		// Act
		err = gen(ctx, destdir, types.KickrGen{Kickr: kickr.Kickr{License: "mit"}})

		// Assert
		require.NoError(t, err)
	})

	t.Run("success_download", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		httpmock.RegisterResponderWithQuery(http.MethodGet, url,
			map[string]string{"fullname": "name", "project": "kickr"},
			httpmock.NewJsonResponderOrPanic(http.StatusOK, gitlab.LicenseTemplate{Content: "some content"}))

		config := types.KickrGen{
			Kickr: kickr.Kickr{
				License:     "mit",
				Maintainers: []*kickr.Maintainer{{Name: "name"}},
			},
			VCS: parser.VCS{ProjectName: "kickr"},
		}

		// Act
		err := gen(ctx, destdir, config)

		// Assert
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(destdir, generator.FileLicense))
		require.NoError(t, err)
		assert.Equal(t, "some content", string(content))
	})
}
