package generate_test

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/generator"
	"github.com/kickr-dev/engine/pkg/parser"
	compare "github.com/kilianpaquier/compare/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

const codeOfConduct = `## Reporting an Issue

Tensions can occur between community members even when they are trying their best to collaborate. Not every conflict represents a code of conduct violation, and this Code of Conduct reinforces encouraged behaviors and norms that can help avoid conflicts and minimize harm.

When an incident does occur, it is important to report it promptly. To report a possible violation, **[NOTE: describe your means of reporting here.]**

Community Moderators take reports of violations seriously and will make every effort to respond in a timely manner. They will investigate all reports of code of conduct violations, reviewing messages, logs, and recordings, or interviewing witnesses and other participants. Community Moderators will keep investigation and enforcement actions as transparent as possible while prioritizing safety and confidentiality. In order to honor these values, enforcement actions are carried out in private with the involved parties, but communicating to the whole community may be part of a mutually agreed upon resolution.


## Addressing and Repairing Harm

**[NOTE: The remedies and repairs outlined below are suggestions based on best practices in code of conduct enforcement. If your community has its own established enforcement process, be sure to edit this section to describe your own policies.]**

If an investigation by the Community Moderators finds that this Code of Conduct has been violated, the following enforcement ladder may be used to determine how best to repair harm, based on the incident's impact on the individuals involved and the community as a whole. Depending on the severity of a violation, lower rungs on the ladder may be skipped.`

func TestGeneratorCodeOfConduct_Exclude(t *testing.T) {
	ctx := t.Context()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	gen := generate.GeneratorCodeOfConduct(http.DefaultClient)

	repo := types.Repository{Config: kickr.Kickr{Exclude: []string{kickr.ExcludeCodeOfConduct}}}

	t.Run("error_remove_code_of_conduct", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		dest := filepath.Join(destdir, generator.FileCodeOfConduct)
		require.NoError(t, os.MkdirAll(filepath.Join(dest, "file.txt"), files.RwxRxRxRx))

		// Act
		err := gen(ctx, destdir, repo)

		// Assert
		assert.ErrorContains(t, err, "remove 'CODE_OF_CONDUCT.md'")
	})

	t.Run("success_remove_existing_code_of_conduct", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()
		dest := filepath.Join(destdir, generator.FileCodeOfConduct)
		f, err := os.Create(dest)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		// Act
		err = gen(ctx, destdir, repo)

		// Assert
		require.NoError(t, err)
		assert.NoFileExists(t, dest)
	})
}

func TestGeneratorCodeOfConduct_Download(t *testing.T) {
	ctx := t.Context()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)
	gen := generate.GeneratorCodeOfConduct(http.DefaultClient)

	t.Run("error_http_call", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		httpmock.RegisterResponder(http.MethodGet, generator.CodeOfConductURL,
			httpmock.NewStringResponder(http.StatusInternalServerError, "error message"))

		// Act
		err := gen(ctx, t.TempDir(), types.Repository{})

		// Assert
		assert.ErrorContains(t, err, "fetch code of conduct")
	})

	t.Run("success_already_exists", func(t *testing.T) {
		// Arrange
		forced := engine.Forced()
		engine.Configure(engine.WithForce(false), engine.WithLogger(engine.GetLogger()))
		t.Cleanup(func() { engine.Configure(engine.WithForce(forced), engine.WithLogger(engine.GetLogger())) })

		destdir := t.TempDir()
		dest := filepath.Join(destdir, generator.FileCodeOfConduct)
		f, err := os.Create(dest)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		// Act
		err = gen(ctx, destdir, types.Repository{})

		// Assert
		require.NoError(t, err)
	})

	t.Run("success_maintainer_and_issue_means", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		t.Cleanup(httpmock.Reset)
		httpmock.RegisterResponder(http.MethodGet, generator.CodeOfConductURL,
			httpmock.NewStringResponder(http.StatusOK, codeOfConduct))

		repo := types.Repository{
			Config: kickr.Kickr{
				Maintainers: []*kickr.Maintainer{
					{Name: "name", Email: "maintainer@example.com"},
				},
			},
			VCS: parser.VCS{
				Platform:    kickr.PlatformGitHub,
				ProjectHost: kickr.PlatformGitHub + ".com",
				ProjectPath: "kickr-dev/kickr",
			},
		}

		// Act
		err := gen(ctx, destdir, repo)

		// Assert
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(destdir, generator.FileCodeOfConduct))
		require.NoError(t, err)
		assert.Equal(t, `<!-- Code generated by kickr; DO NOT EDIT. -->

<!-- Created by https://www.contributor-covenant.org/version/3/0/code_of_conduct/code_of_conduct.md -->

## Reporting an Issue

Tensions can occur between community members even when they are trying their best to collaborate. Not every conflict represents a code of conduct violation, and this Code of Conduct reinforces encouraged behaviors and norms that can help avoid conflicts and minimize harm.

When an incident does occur, it is important to report it promptly. To report a possible violation, use one of the following means:

- open an [issue](https://github.com/kickr-dev/kickr/issues)
- contact [name](mailto:maintainer@example.com)

Community Moderators take reports of violations seriously and will make every effort to respond in a timely manner. They will investigate all reports of code of conduct violations, reviewing messages, logs, and recordings, or interviewing witnesses and other participants. Community Moderators will keep investigation and enforcement actions as transparent as possible while prioritizing safety and confidentiality. In order to honor these values, enforcement actions are carried out in private with the involved parties, but communicating to the whole community may be part of a mutually agreed upon resolution.


## Addressing and Repairing Harm

If an investigation by the Community Moderators finds that this Code of Conduct has been violated, the following enforcement ladder may be used to determine how best to repair harm, based on the incident's impact on the individuals involved and the community as a whole. Depending on the severity of a violation, lower rungs on the ladder may be skipped.

<!-- End of https://www.contributor-covenant.org/version/3/0/code_of_conduct/code_of_conduct.md -->
`, string(bytes.ReplaceAll(content, compare.Carriage, []byte{})))
	})

	t.Run("success_no_maintainer_issue_only", func(t *testing.T) {
		// Arrange
		destdir := t.TempDir()

		t.Cleanup(httpmock.Reset)
		httpmock.RegisterResponder(http.MethodGet, generator.CodeOfConductURL,
			httpmock.NewStringResponder(http.StatusOK, codeOfConduct))

		repo := types.Repository{
			VCS: parser.VCS{
				Platform:    kickr.PlatformGitLab,
				ProjectHost: kickr.PlatformGitLab + ".com",
				ProjectPath: "kickr-dev/kickr",
			},
		}

		// Act
		err := gen(ctx, destdir, repo)

		// Assert
		require.NoError(t, err)
		content, err := os.ReadFile(filepath.Join(destdir, generator.FileCodeOfConduct))
		require.NoError(t, err)
		assert.Equal(t, `<!-- Code generated by kickr; DO NOT EDIT. -->

<!-- Created by https://www.contributor-covenant.org/version/3/0/code_of_conduct/code_of_conduct.md -->

## Reporting an Issue

Tensions can occur between community members even when they are trying their best to collaborate. Not every conflict represents a code of conduct violation, and this Code of Conduct reinforces encouraged behaviors and norms that can help avoid conflicts and minimize harm.

When an incident does occur, it is important to report it promptly. To report a possible violation, use one of the following means:

- open an [issue](https://gitlab.com/kickr-dev/kickr/-/work_items)

Community Moderators take reports of violations seriously and will make every effort to respond in a timely manner. They will investigate all reports of code of conduct violations, reviewing messages, logs, and recordings, or interviewing witnesses and other participants. Community Moderators will keep investigation and enforcement actions as transparent as possible while prioritizing safety and confidentiality. In order to honor these values, enforcement actions are carried out in private with the involved parties, but communicating to the whole community may be part of a mutually agreed upon resolution.


## Addressing and Repairing Harm

If an investigation by the Community Moderators finds that this Code of Conduct has been violated, the following enforcement ladder may be used to determine how best to repair harm, based on the incident's impact on the individuals involved and the community as a whole. Depending on the severity of a violation, lower rungs on the ladder may be skipped.

<!-- End of https://www.contributor-covenant.org/version/3/0/code_of_conduct/code_of_conduct.md -->
`, string(bytes.ReplaceAll(content, compare.Carriage, []byte{})))
	})
}
