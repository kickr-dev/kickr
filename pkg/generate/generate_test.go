package generate_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/files"
	"github.com/kickr-dev/engine/pkg/parser"
	compare "github.com/kilianpaquier/compare/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
	"github.com/kickr-dev/kickr/testutils"
)

func TestGenerate_NoLang(t *testing.T) {
	ctx := t.Context()

	t.Run("success_chart", func(t *testing.T) {
		// Arrange
		cases := []kickr.CI{
			{Provider: parser.GitHub, Helm: &kickr.Helm{}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmNone}},

			{Provider: parser.GitLab, Helm: &kickr.Helm{}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmNone}},
		}

		for _, ci := range cases {
			publish := "nil"
			if ci.Helm != nil && ci.Helm.Publish != "" {
				publish = ci.Helm.Publish
			}
			name := fmt.Sprint(ci.Provider, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("success_renovate", func(t *testing.T) {
		type testcase struct {
			CI   string
			Auth string
		}

		cases := []testcase{
			{CI: parser.GitHub},

			{CI: parser.GitHub, Auth: kickr.AuthGitHubApp},
			{CI: parser.GitHub, Auth: kickr.AuthGitHubToken},
			{CI: parser.GitHub, Auth: kickr.AuthPersonalToken},

			{CI: parser.GitLab},
		}
		for _, tc := range cases {
			name := tc.CI
			if tc.Auth != "" {
				name += "_auth_" + tc.Auth
			}

			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						Dependencies: &kickr.Dependencies{Manager: kickr.ManagerRenovate, Local: "configs/renovate.json5"},
						CI:           &kickr.CI{Provider: tc.CI, Renovate: &kickr.Renovate{Auth: tc.Auth}},
						Exclude:      []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
						Platform:     tc.CI,
					},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}

		t.Run("templates", func(t *testing.T) {
			// Arrange
			tmpl := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
				file, err := os.Create(filepath.Join(destdir, "template.tmpl"))
				if err != nil {
					return fmt.Errorf("create: %w", err)
				}
				return file.Close()
			}

			config := types.KickrWrapper{
				Kickr: kickr.Kickr{
					Dependencies: &kickr.Dependencies{Manager: kickr.ManagerRenovate},
					Exclude:      []string{kickr.ExcludeMakefile},
				},
			}

			// Act & Assert
			test(ctx, t, config, tmpl)
		})
	})

	t.Run("success_precommit", func(t *testing.T) {
		for _, precommit := range []bool{true, false} {
			t.Run(strconv.FormatBool(precommit), func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:      &kickr.CI{Provider: parser.GitHub},
						Exclude: []string{kickr.ExcludeMakefile},
					},
				}
				if !precommit {
					config.Exclude = append(config.Exclude, kickr.ExcludePreCommit)
				} else {
					config.CI.Options = append(config.CI.Options, kickr.OptionPreCommitAutoCommit)
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("success_release", func(t *testing.T) {
		type testcase struct {
			Auth string
			Auto bool
			CI   string
		}

		cases := []testcase{
			{CI: parser.GitHub},
			{CI: parser.GitHub, Auto: true},

			{CI: parser.GitHub, Auth: kickr.AuthGitHubApp},
			{CI: parser.GitHub, Auth: kickr.AuthGitHubToken},
			{CI: parser.GitHub, Auth: kickr.AuthPersonalToken},

			{CI: parser.GitLab},
			{CI: parser.GitLab, Auto: true},
		}
		for _, tc := range cases {
			name := tc.CI
			if tc.Auto {
				name += "_auto"
			}
			if tc.Auth != "" {
				name += "_auth_" + tc.Auth
			}

			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.CI,
							Release:  &kickr.Release{Auto: tc.Auto, Auth: tc.Auth},
						},
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: tc.CI,
					},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})
}

func TestGenerate_Shell(t *testing.T) {
	ctx := t.Context()

	shell := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
		return os.WriteFile(filepath.Join(destdir, "script.sh"), []byte("#!/bin/sh\n"), files.RwxRxRxRx)
	}

	t.Run("success_ci", func(t *testing.T) {
		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:      &kickr.CI{Provider: ci},
						Exclude: []string{kickr.ExcludeMakefile},
					},
				}

				// Act & Assert
				test(ctx, t, config, shell)
			})
		}
	})

	t.Run("success_precommit", func(t *testing.T) {
		for _, precommit := range []bool{true, false} {
			t.Run(strconv.FormatBool(precommit), func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{Kickr: kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile}}}
				if !precommit {
					config.Exclude = append(config.Exclude, kickr.ExcludePreCommit)
				}

				// Act & Assert
				test(ctx, t, config, shell)
			})
		}
	})
}

func TestGenerate_Golang(t *testing.T) {
	ctx := t.Context()

	t.Run("success_cli", func(t *testing.T) {
		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						Dependencies: &kickr.Dependencies{Manager: kickr.ManagerDependabot},
						CI:           &kickr.CI{Provider: ci, Release: &kickr.Release{}},
						Platform:     ci,
					},
				}
				golang := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.AddCLI("name")

					gomod := parser.Gomod{
						Go:     "1.23",
						Module: ci + ".com/kickr-dev/kickr",
					}
					config.VCS = gomod.AsVCS()
					config.SetLanguage("go", gomod)
					return nil
				}

				// Act & Assert
				test(ctx, t, config, golang)
			})
		}
	})

	t.Run("success_library", func(t *testing.T) {
		for _, platform := range []string{parser.GitLab, parser.GitHub} {
			t.Run(platform, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile},
						PreCommit: []string{kickr.PreCommitGomodTidy},
						Platform:  platform,
					},
				}
				golang := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					gomod := parser.Gomod{
						Go:     "1.23",
						Module: platform + ".com/kickr-dev/kickr",
					}
					config.VCS = gomod.AsVCS()
					config.SetLanguage("go", gomod)
					return nil
				}

				// Act & Assert
				test(ctx, t, config, golang)
			})
		}
	})

	t.Run("success_multiple_bin_helm", func(t *testing.T) {
		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						Dependencies: &kickr.Dependencies{Manager: kickr.ManagerRenovate},
						CI: &kickr.CI{
							Provider: ci,
							Docker:   &kickr.Docker{Path: "path/to/registry", Registry: "registry.example.com"},
							Helm:     &kickr.Helm{Deploy: kickr.HelmManual, Path: "path/to/repository", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"},
							Options:  []string{kickr.OptionCodeCov, kickr.OptionCodeQL, kickr.OptionHardenRunner, kickr.OptionLabeler, kickr.OptionScoreCardOSSF, kickr.OptionSonarQube},
							Release:  &kickr.Release{},
						},
						Description: "A useful project description",
						Exclude:     []string{kickr.ExcludeShell},
						Platform:    ci,
					},
				}
				golang := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.AddJob("job-name")
					config.AddCron("cron-name")
					config.AddWorker("worker-name")

					gomod := parser.Gomod{
						Go:     "1.23",
						Module: ci + ".com/kickr-dev/kickr",
					}
					config.VCS = gomod.AsVCS()
					config.SetLanguage("go", gomod)
					return nil
				}

				// Act & Assert
				test(ctx, t, config, golang)
			})
		}
	})
}

func TestGenerate_Hugo(t *testing.T) {
	ctx := t.Context()

	t.Run("success_netlify", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Netlify: &kickr.Netlify{Auto: true}},
			{Provider: parser.GitHub, Netlify: &kickr.Netlify{}},

			{Provider: parser.GitLab, Netlify: &kickr.Netlify{Auto: true}},
			{Provider: parser.GitLab, Netlify: &kickr.Netlify{}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Netlify.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{CI: &ci, Platform: ci.Provider},
				}
				hugo := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.SetLanguage("hugo", nil)
					return nil
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})

	t.Run("success_pages", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Pages: &kickr.Pages{Auto: true}},
			{Provider: parser.GitHub, Pages: &kickr.Pages{}},

			{Provider: parser.GitLab, Pages: &kickr.Pages{Auto: true}},
			{Provider: parser.GitLab, Pages: &kickr.Pages{}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Pages.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{CI: &ci, Platform: ci.Provider},
				}
				hugo := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.SetLanguage("hugo", nil)
					return nil
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})
}

func TestGenerate_Node(t *testing.T) {
	ctx := t.Context()

	t.Run("success_package_managers", func(t *testing.T) {
		for _, tc := range []string{"bun@1.1.6", "npm@7.0.0", "pnpm@9.0.0", "yarn@1.22.10"} {
			t.Run(tc, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: parser.GitHub},
						Platform: parser.GitHub,
					},
				}
				node := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.AddWorker("index.js")
					config.SetLanguage("node", parser.PackageJSON{Name: "kickr", PackageManager: tc})
					return nil
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_library", func(t *testing.T) {
		type testcase struct {
			Manager        string
			CI             string
			PackageManager string
		}
		cases := []testcase{
			{Manager: kickr.ManagerRenovate, CI: parser.GitHub, PackageManager: "bun@1.1.6"},
			{Manager: kickr.ManagerDependabot, CI: parser.GitHub, PackageManager: "bun@1.1.6"},
			{Manager: kickr.ManagerRenovate, CI: parser.GitHub, PackageManager: "npm@7.0.0"},
			{Manager: kickr.ManagerDependabot, CI: parser.GitHub, PackageManager: "npm@7.0.0"},

			{Manager: kickr.ManagerRenovate, CI: parser.GitLab, PackageManager: "bun@1.1.6"},
			{Manager: kickr.ManagerDependabot, CI: parser.GitLab, PackageManager: "bun@1.1.6"},
			{Manager: kickr.ManagerRenovate, CI: parser.GitLab, PackageManager: "npm@7.0.0"},
			{Manager: kickr.ManagerDependabot, CI: parser.GitLab, PackageManager: "npm@7.0.0"},
		}

		for _, tc := range cases {
			name := fmt.Sprint(tc.CI, "_", tc.Manager, "_", tc.PackageManager)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						Dependencies: &kickr.Dependencies{Manager: tc.Manager},
						CI: &kickr.CI{
							Provider: tc.CI,
							Release:  &kickr.Release{Options: []string{kickr.OptionBackmerge}},
							Renovate: &kickr.Renovate{Auth: kickr.AuthPersonalToken},
						},
						Platform: tc.CI,
					},
				}
				node := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.SetLanguage("node", parser.PackageJSON{Name: "kickr", PackageManager: tc.PackageManager})
					return nil
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_netlify", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Netlify: &kickr.Netlify{Auto: true}},
			{Provider: parser.GitHub, Netlify: &kickr.Netlify{}},

			{Provider: parser.GitLab, Netlify: &kickr.Netlify{Auto: true}},
			{Provider: parser.GitLab, Netlify: &kickr.Netlify{}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Netlify.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
					},
				}
				node := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.AddWorker("index.js")
					config.SetLanguage("node", parser.PackageJSON{Name: "kickr", PackageManager: "bun@1.1.6"})
					return nil
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_pages", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Pages: &kickr.Pages{Auto: true}},
			{Provider: parser.GitHub, Pages: &kickr.Pages{}},

			{Provider: parser.GitLab, Pages: &kickr.Pages{Auto: true}},
			{Provider: parser.GitLab, Pages: &kickr.Pages{}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Pages.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
					},
				}
				node := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.AddWorker("index.js")
					config.SetLanguage("node", parser.PackageJSON{Name: "kickr", PackageManager: "bun@1.1.6"})
					return nil
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_helm", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Helm: &kickr.Helm{}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Deploy: kickr.HelmAuto}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Deploy: kickr.HelmManual}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmManual}},

			{Provider: parser.GitLab, Helm: &kickr.Helm{}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Deploy: kickr.HelmAuto}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Deploy: kickr.HelmManual}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Provider: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmManual}},
		}
		for _, ci := range cases {
			publish := kickr.HelmNone
			if ci.Helm.Publish != "" {
				publish = ci.Helm.Publish
			}
			deploy := kickr.HelmNone
			if ci.Helm.Deploy != "" {
				deploy = ci.Helm.Deploy
			}

			name := fmt.Sprint(ci.Provider, "_deploy_", deploy, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:       &ci,
						Exclude:  []string{kickr.ExcludeMakefile},
						Platform: ci.Provider,
					},
				}
				node := func(_ context.Context, _ string, config *types.KickrWrapper) error {
					config.AddWorker("index.js")
					config.SetLanguage("node", parser.PackageJSON{Name: "kickr", PackageManager: "bun@1.1.6"})
					return nil
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})
}

func ParserInfo(_ context.Context, _ string, config *types.KickrWrapper) error {
	config.VCS = parser.VCS{
		Platform:    config.Platform,
		ProjectHost: config.Platform + ".com",
		ProjectName: "kickr",
		ProjectPath: "kickr-dev/kickr",
	}
	return nil
}

// test verifies every generation with provided config, parser and t.Name folder expected results.
func test(ctx context.Context, t *testing.T, config types.KickrWrapper, parsers ...engine.Parser[types.KickrWrapper]) {
	t.Helper()

	// Arrange
	config.Maintainers = append(config.Maintainers, &kickr.Maintainer{Name: "kilianpaquier"})
	assertdir := filepath.Join(testutils.Testdata(t), t.Name())
	require.NoError(t, os.MkdirAll(assertdir, files.RwxRxRxRx))

	destdir := t.TempDir()
	if ok, _ := strconv.ParseBool(os.Getenv("TESTDATA")); ok {
		destdir = assertdir
	}

	// Act
	_, err := engine.Generate(ctx, destdir, config,
		slices.Concat(parsers, []engine.Parser[types.KickrWrapper]{ParserInfo, generate.ParserGlob, generate.ParserGolang, generate.ParserNode, generate.ParserHelm}),
		[]engine.Generator[types.KickrWrapper]{
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.Dependabot(), templates.Renovate())),
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.CodeCov(), templates.Sonar())),
			engine.GeneratorTemplates(templates.FS(), templates.Docker()),
			engine.GeneratorTemplates(templates.FS(), templates.Golang()),
			engine.GeneratorTemplates(templates.FS(), templates.Misc()),
			engine.GeneratorTemplates(templates.FS(), templates.Makefile()),
			engine.GeneratorTemplates(templates.FS(), templates.Chart()),
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.GitHub(), templates.GitLab(), templates.SemanticRelease())),
		})

	// Assert
	require.NoError(t, err)
	assert.NoError(t, compare.Dirs(assertdir, destdir))
}
