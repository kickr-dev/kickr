package helm_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	compare "github.com/kilianpaquier/compare/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yannh/kubeconform/pkg/validator"
	"helm.sh/helm/v4/pkg/chart/common"
	"helm.sh/helm/v4/pkg/chart/common/util"
	"helm.sh/helm/v4/pkg/chart/v2/lint"
	"helm.sh/helm/v4/pkg/chart/v2/lint/support"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
	helmengine "helm.sh/helm/v4/pkg/engine"
	releaseutil "helm.sh/helm/v4/pkg/release/v1/util"

	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
	"github.com/kickr-dev/kickr/testutils"
)

const kubernetesVersion = "1.37.0"

func TestHelmTemplate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	testdir := filepath.Join(testutils.Testdata(t), t.Name())
	generate := engine.GeneratorTemplates(templates.FS(), templates.Chart()) // chart generation

	// read all tests (simpler in case new test cases would be added)
	tests, err := os.ReadDir(testdir)
	require.NoError(t, err)

	// parse kube version since lint doesn't take latest one
	kubeVersion, err := common.ParseKubeVersion(kubernetesVersion)
	require.NoError(t, err)

	// share one validator to reuse its in-memory schemas, the disk cache avoids downloading them again on next runs
	cachedir, err := os.UserCacheDir()
	require.NoError(t, err)

	kubeconform := filepath.Join(cachedir, "kickr", "kubeconform")
	require.NoError(t, os.MkdirAll(kubeconform, files.RwxRxRxRx))

	v, err := validator.New(nil, validator.Opts{Cache: kubeconform, KubernetesVersion: kubernetesVersion, Strict: true})
	require.NoError(t, err)

	// run all tests
	for _, test := range tests {
		if !test.IsDir() {
			continue
		}

		t.Run(test.Name(), func(t *testing.T) {
			t.Parallel()

			// Arrange
			assertdir := filepath.Join(testdir, test.Name())
			expected := filepath.Join(assertdir, "manifest.yaml")
			actual := filepath.Join(t.TempDir(), "manifest.yaml")

			// generate chart files
			destdir := t.TempDir()
			repo := types.Repository{
				Config: kickr.Kickr{Helm: &kickr.Helm{}},
				Modules: []types.Module{
					{
						Directory: types.RootModule,
						Languages: map[string]any{
							types.LanguageHelm: map[string]any{"projectName": "kickr"},
						},
					},
				},
			}
			require.NoError(t, generate(ctx, testutils.Root(t, destdir), repo))
			chartdir := filepath.Join(destdir, "chart")

			// remove default values since we use custom ones
			require.NoError(t, os.Remove(filepath.Join(chartdir, "values.yaml")))

			// copy chart additional inputs for given test
			chartinput := filepath.Join(assertdir, "chart")
			if files.Exists(chartinput) {
				require.NoError(t, os.CopyFS(chartdir, os.DirFS(chartinput)))
			}

			// Act
			manifest, err := template(ctx, *kubeVersion, chartdir, filepath.Join(assertdir, "values.yaml"))

			// Assert
			require.NoError(t, err, manifest)
			require.NoError(t, os.WriteFile(actual, []byte(manifest), files.RwRR))

			t.Run("lint", func(t *testing.T) {
				t.Parallel()

				// Arrange
				values, err := readValues(filepath.Join(assertdir, "values.yaml"))
				require.NoError(t, err)

				// Act
				result := lint.RunAll(chartdir, values, "", lint.WithKubeVersion(kubeVersion))

				// Assert
				assert.Less(t, result.HighestSeverity, support.ErrorSev, result.Messages)
			})

			t.Run("template", func(t *testing.T) {
				t.Parallel()

				// Assert
				assert.NoError(t, compare.Files(expected, actual))
			})

			t.Run("kubeconform", func(t *testing.T) {
				t.Parallel()

				// Arrange
				file, err := os.Open(actual) // close is handled by ValidateWithContext ...
				require.NoError(t, err)

				// Act
				results := v.ValidateWithContext(ctx, actual, file)

				// Assert
				for _, result := range results {
					assert.NotContains(t, []validator.Status{validator.Invalid, validator.Error}, result.Status, result.Err)
				}
			})
		})
	}
}

// template renders the chart like an helm install client dry run and returns the computed manifest.
//
// On sorting failure, all rendered files are returned alongside the error to help debugging.
func template(ctx context.Context, kubeVersion common.KubeVersion, chartdir, valuesFile string) (string, error) {
	c, err := loader.LoadDir(chartdir)
	if err != nil {
		return "", fmt.Errorf("load chart dir: %w", err)
	}
	values, err := readValues(valuesFile)
	if err != nil {
		return "", err // already wrapped
	}

	if err := chartutil.ProcessDependencies(c, values); err != nil {
		return "", fmt.Errorf("process dependencies: %w", err)
	}

	caps := &common.Capabilities{
		KubeVersion: kubeVersion,
		APIVersions: common.DefaultVersionSet,
	}
	options := common.ReleaseOptions{Name: c.Name(), Revision: 1, IsInstall: true}
	renderValues, err := util.ToRenderValuesWithSchemaValidation(c, values, options, caps, false)
	if err != nil {
		return "", fmt.Errorf("to render values: %w", err)
	}

	rendered, err := helmengine.Engine{}.RenderWithContext(ctx, c, renderValues)
	if err != nil {
		return "", fmt.Errorf("render: %w", err)
	}

	var manifest strings.Builder
	_, manifests, err := releaseutil.SortManifests(rendered, nil, releaseutil.InstallOrder)
	if err != nil {
		for name, content := range rendered {
			if strings.TrimSpace(content) != "" {
				fmt.Fprintf(&manifest, "---\n# Source: %s\n%s\n", name, content)
			}
		}
		return manifest.String(), fmt.Errorf("sort manifests: %w", err)
	}
	for _, m := range manifests {
		fmt.Fprintf(&manifest, "---\n# Source: %s\n%s\n", m.Name, m.Content)
	}
	return manifest.String(), nil
}

// readValues reads the given values file, a missing file means no values.
func readValues(valuesFile string) (map[string]any, error) {
	raw, err := os.ReadFile(valuesFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var values map[string]any
	if err := yaml.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return values, nil
}
