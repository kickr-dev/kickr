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

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
	"github.com/kickr-dev/kickr/pkg/generate"
	"github.com/kickr-dev/kickr/pkg/generate/templates"
	"github.com/kickr-dev/kickr/testutils"
)

func TestGenerate_NoLang(t *testing.T) {
	ctx := t.Context()

	t.Run("success_chart", func(t *testing.T) {
		// Arrange
		cases := []kickr.CI{
			{Name: parser.GitHub, Helm: &kickr.Helm{}},
			{Name: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Name: parser.GitHub, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"}},
			{Name: parser.GitHub, Helm: &kickr.Helm{Publish: kickr.HelmNone}},

			{Name: parser.GitLab, Helm: &kickr.Helm{}},
			{Name: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmAuto}},
			{Name: parser.GitLab, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"}},
			{Name: parser.GitLab, Helm: &kickr.Helm{Publish: kickr.HelmNone}},
		}

		for _, ci := range cases {
			publish := "nil"
			if ci.Helm != nil && ci.Helm.Publish != "" {
				publish = ci.Helm.Publish
			}
			name := fmt.Sprint(ci.Name, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					CI:      &ci,
					Exclude: []string{kickr.Makefile},
					VCS:     parser.VCS{Platform: ci.Name},
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

			{CI: parser.GitHub, Auth: kickr.GitHubApp},
			{CI: parser.GitHub, Auth: kickr.GitHubToken},
			{CI: parser.GitHub, Auth: kickr.PersonalToken},

			{CI: parser.GitLab},
		}
		for _, tc := range cases {
			name := tc.CI
			if tc.Auth != "" {
				name += "_auth_" + tc.Auth
			}

			t.Run(name, func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					Bot:     kickr.Renovate,
					CI:      &kickr.CI{Auth: kickr.Auth{Maintenance: tc.Auth}, Name: tc.CI},
					Exclude: []string{kickr.Makefile, kickr.Shell},
					Include: []string{kickr.RenovatePostUpgrade},
					VCS:     parser.VCS{Platform: tc.CI},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}

		t.Run("templates", func(t *testing.T) {
			// Arrange
			tmpl := func(_ context.Context, destdir string, _ *kickr.Config) error {
				file, err := os.Create(filepath.Join(destdir, "template.tmpl"))
				if err != nil {
					return fmt.Errorf("create: %w", err)
				}
				return file.Close()
			}

			config := kickr.Config{
				Bot:     kickr.Renovate,
				Exclude: []string{kickr.Makefile},
			}

			// Act & Assert
			test(ctx, t, config, tmpl)
		})
	})

	t.Run("success_precommit", func(t *testing.T) {
		for _, precommit := range []bool{true, false} {
			t.Run(strconv.FormatBool(precommit), func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					CI:      &kickr.CI{Name: parser.GitHub},
					Exclude: []string{kickr.Makefile},
				}
				if !precommit {
					config.Exclude = append(config.Exclude, kickr.PreCommit)
				} else {
					config.CI.Options = append(config.CI.Options, kickr.PreCommitAutoCommit)
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

			{CI: parser.GitHub, Auth: kickr.GitHubApp},
			{CI: parser.GitHub, Auth: kickr.GitHubToken},
			{CI: parser.GitHub, Auth: kickr.PersonalToken},

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
				config := kickr.Config{
					CI: &kickr.CI{
						Auth:    kickr.Auth{Release: tc.Auth},
						Name:    tc.CI,
						Release: &kickr.Release{Auto: tc.Auto},
					},
					Exclude: []string{kickr.Makefile},
					VCS:     parser.VCS{Platform: tc.CI},
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})
}

func TestGenerate_Shell(t *testing.T) {
	ctx := t.Context()

	shell := func(_ context.Context, destdir string, _ *kickr.Config) error {
		return os.WriteFile(filepath.Join(destdir, "script.sh"), []byte("#!/bin/sh\n"), files.RwxRxRxRx)
	}

	t.Run("success_ci", func(t *testing.T) {
		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					CI:      &kickr.CI{Name: ci},
					Exclude: []string{kickr.Makefile},
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
				config := kickr.Config{Exclude: []string{kickr.Makefile}}
				if !precommit {
					config.Exclude = append(config.Exclude, kickr.PreCommit)
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
				config := kickr.Config{
					Bot: kickr.Dependabot,
					CI:  &kickr.CI{Name: ci, Release: &kickr.Release{}},
					VCS: parser.VCS{Platform: ci},
				}
				golang := func(_ context.Context, _ string, config *kickr.Config) error {
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
				config := kickr.Config{
					Exclude: []string{kickr.Makefile},
					Include: []string{kickr.PreCommitGomodTidy},
					VCS:     parser.VCS{Platform: platform},
				}
				golang := func(_ context.Context, _ string, config *kickr.Config) error {
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
				config := kickr.Config{
					CI: &kickr.CI{
						Name:       ci,
						Deployment: &kickr.Deployment{Platform: kickr.Kubernetes},
						Docker:     &kickr.Docker{Path: "path/to/registry", Registry: "registry.example.com"},
						Helm:       &kickr.Helm{Path: "path/to/repository", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"},
						Options:    []string{kickr.CodeCov, kickr.CodeQL, kickr.Sonar, kickr.Labeler},
						Release:    &kickr.Release{},
					},
					Description: "A useful project description",
					Exclude:     []string{kickr.Shell},
					VCS:         parser.VCS{Platform: ci},
				}
				golang := func(_ context.Context, _ string, config *kickr.Config) error {
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

	cases := []kickr.CI{
		{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Netlify, Auto: true}},
		{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Netlify}},
		{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Pages, Auto: true}},
		{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Pages}},

		{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Netlify, Auto: true}},
		{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Netlify}},
		{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Pages, Auto: true}},
		{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Pages}},
	}
	for _, ci := range cases {
		name := fmt.Sprint(ci.Name, "_", ci.Deployment.Platform, "_auto_", ci.Deployment.Auto)
		t.Run(name, func(t *testing.T) {
			// Arrange
			config := kickr.Config{
				CI:  &ci,
				VCS: parser.VCS{Platform: ci.Name},
			}
			hugo := func(_ context.Context, _ string, config *kickr.Config) error {
				config.SetLanguage("hugo", nil)
				return nil
			}

			// Act & Assert
			test(ctx, t, config, hugo)
		})
	}
}

func TestGenerate_Node(t *testing.T) {
	ctx := t.Context()

	t.Run("success_package_managers", func(t *testing.T) {
		for _, tc := range []string{"bun@1.1.6", "npm@7.0.0", "pnpm@9.0.0", "yarn@1.22.10"} {
			t.Run(tc, func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					CI:  &kickr.CI{Name: parser.GitHub},
					VCS: parser.VCS{Platform: parser.GitHub},
				}
				node := func(_ context.Context, _ string, config *kickr.Config) error {
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
			Bot            string
			CI             string
			PackageManager string
		}
		cases := []testcase{
			{Bot: kickr.Renovate, CI: parser.GitLab, PackageManager: "bun@1.1.6"},
			{Bot: kickr.Dependabot, CI: parser.GitLab, PackageManager: "bun@1.1.6"},
			{Bot: kickr.Renovate, CI: parser.GitLab, PackageManager: "npm@7.0.0"},
			{Bot: kickr.Dependabot, CI: parser.GitLab, PackageManager: "npm@7.0.0"},

			{Bot: kickr.Renovate, CI: parser.GitHub, PackageManager: "bun@1.1.6"},
			{Bot: kickr.Dependabot, CI: parser.GitHub, PackageManager: "bun@1.1.6"},
			{Bot: kickr.Renovate, CI: parser.GitHub, PackageManager: "npm@7.0.0"},
			{Bot: kickr.Dependabot, CI: parser.GitHub, PackageManager: "npm@7.0.0"},
		}

		for _, tc := range cases {
			name := fmt.Sprint(tc.CI, "_", tc.Bot, "_", tc.PackageManager)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					Bot: tc.Bot,
					CI: &kickr.CI{
						Name:    tc.CI,
						Auth:    kickr.Auth{Maintenance: kickr.PersonalToken},
						Release: &kickr.Release{Backmerge: true},
					},
					VCS: parser.VCS{Platform: tc.CI},
				}
				node := func(_ context.Context, _ string, config *kickr.Config) error {
					config.SetLanguage("node", parser.PackageJSON{Name: "kickr", PackageManager: tc.PackageManager})
					return nil
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_deployment", func(t *testing.T) {
		statics := []kickr.CI{
			{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Netlify, Auto: true}},
			{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Netlify}},
			{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Pages, Auto: true}},
			{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Pages}},

			{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Netlify, Auto: true}},
			{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Netlify}},
			{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Pages, Auto: true}},
			{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Pages}},
		}
		for _, ci := range statics {
			name := fmt.Sprint(ci.Name, "_", ci.Deployment.Platform, "_auto_", ci.Deployment.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					CI:      &ci,
					Exclude: []string{kickr.Makefile},
					VCS:     parser.VCS{Platform: ci.Name},
				}
				node := func(_ context.Context, _ string, config *kickr.Config) error {
					config.AddWorker("index.js")
					config.SetLanguage("node", parser.PackageJSON{Name: "kickr", PackageManager: "bun@1.1.6"})
					return nil
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}

		cases := []kickr.CI{
			{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Kubernetes}, Helm: &kickr.Helm{}},
			{Name: parser.GitHub, Deployment: &kickr.Deployment{Platform: kickr.Kubernetes}, Helm: &kickr.Helm{Publish: kickr.HelmManual}},

			{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Kubernetes}, Helm: &kickr.Helm{}},
			{Name: parser.GitLab, Deployment: &kickr.Deployment{Platform: kickr.Kubernetes}, Helm: &kickr.Helm{Publish: kickr.HelmManual}},
		}
		for _, ci := range cases {
			publish := kickr.HelmNone
			if ci.Helm != nil && ci.Helm.Publish != "" {
				publish = ci.Helm.Publish
			}
			name := fmt.Sprint(ci.Name, "_", ci.Deployment.Platform, "_helm_", ci.Helm != nil, "_publish_", publish)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := kickr.Config{
					CI:      &ci,
					Exclude: []string{kickr.Makefile},
					VCS:     parser.VCS{Platform: ci.Name},
				}
				node := func(_ context.Context, _ string, config *kickr.Config) error {
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

func ParserInfo(_ context.Context, _ string, config *kickr.Config) error {
	config.VCS = parser.VCS{
		Platform:    config.Platform,
		ProjectHost: config.Platform + ".com",
		ProjectName: "kickr",
		ProjectPath: "kickr-dev/kickr",
	}
	return nil
}

// test verifies every generation with provided config, parser and t.Name folder expected results.
func test(ctx context.Context, t *testing.T, config kickr.Config, parsers ...engine.Parser[kickr.Config]) {
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
		slices.Concat(parsers, []engine.Parser[kickr.Config]{ParserInfo, generate.ParserGolang, generate.ParserNode, generate.ParserHelm}),
		[]engine.Generator[kickr.Config]{
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
