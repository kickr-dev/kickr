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
						CI:           &kickr.CI{Provider: tc.CI, Renovate: &kickr.Renovate{Auth: tc.Auth}},
						Dependencies: &kickr.Dependencies{Manager: kickr.ManagerRenovate, Local: "configs/renovate.json5"},
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
					config.PreCommit = append(config.PreCommit, kickr.PreCommitAutoCommit)
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
		// Arrange
		golang := func(ci string) func(ctx context.Context, destdir string, config *types.KickrWrapper) error {
			return func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\ngo 1.23\n", ci)
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}

				cmd := filepath.Join(destdir, parser.FolderCMD)
				if err := os.MkdirAll(cmd, files.RwxRxRxRx); err != nil {
					return fmt.Errorf("mkdir all: %w", err)
				}
				for _, bin := range []string{"name"} {
					if err := os.MkdirAll(filepath.Join(cmd, bin), files.RwxRxRxRx); err != nil {
						return fmt.Errorf("mkdir all: %w", err)
					}
					file, err := os.Create(filepath.Join(cmd, bin, parser.FileMain))
					if err != nil {
						return fmt.Errorf("create: %w", err)
					}
					if err := file.Close(); err != nil {
						return fmt.Errorf("close: %w", err)
					}
				}

				return nil
			}
		}

		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:           &kickr.CI{Provider: ci, Release: &kickr.Release{}},
						Dependencies: &kickr.Dependencies{Manager: kickr.ManagerDependabot},
						Platform:     ci,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(ci))
			})
		}
	})

	t.Run("success_library", func(t *testing.T) {
		// Arrange
		golang := func(platform string) func(ctx context.Context, destdir string, config *types.KickrWrapper) error {
			return func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\ngo 1.23\n", platform)
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}
				return nil
			}
		}

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

				// Act & Assert
				test(ctx, t, config, golang(platform))
			})
		}
	})

	t.Run("success_multiple_bin_helm", func(t *testing.T) {
		// Arrange
		golang := func(ci string) func(ctx context.Context, destdir string, config *types.KickrWrapper) error {
			return func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\ngo 1.23\n", ci)
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}

				cmd := filepath.Join(destdir, parser.FolderCMD)
				if err := os.MkdirAll(cmd, files.RwxRxRxRx); err != nil {
					return fmt.Errorf("mkdir all: %w", err)
				}
				for _, bin := range []string{"cron-name", "job-name", "worker-name"} {
					if err := os.MkdirAll(filepath.Join(cmd, bin), files.RwxRxRxRx); err != nil {
						return fmt.Errorf("mkdir all: %w", err)
					}
					file, err := os.Create(filepath.Join(cmd, bin, parser.FileMain))
					if err != nil {
						return fmt.Errorf("create: %w", err)
					}
					if err := file.Close(); err != nil {
						return fmt.Errorf("close: %w", err)
					}
				}

				return nil
			}
		}

		for _, ci := range []string{parser.GitLab, parser.GitHub} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: ci,
							Docker:   &kickr.Docker{Path: "path/to/registry", Registry: "registry.example.com"},
							Helm:     &kickr.Helm{Deploy: kickr.HelmManual, Path: "path/to/repository", Publish: kickr.HelmManual, Registry: "chartmuseum.example.com"},
							Options:  []string{kickr.OptionCodeCov, kickr.OptionCodeQL, kickr.OptionHardenRunner, kickr.OptionLabeler, kickr.OptionScoreCardOSSF, kickr.OptionSonarQube},
							Release:  &kickr.Release{},
						},
						Description:  "A useful project description",
						Dependencies: &kickr.Dependencies{Manager: kickr.ManagerRenovate},
						Exclude:      []string{kickr.ExcludeShell},
						Platform:     ci,
						PreCommit:    []string{kickr.PreCommitGolangciLint},
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(ci))
			})
		}
	})
}

func TestGenerate_Hugo(t *testing.T) {
	ctx := t.Context()

	hugo := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
		file, err := os.Create(filepath.Join(destdir, "hugo.toml"))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return file.Close()
	}

	t.Run("success_no_website", func(t *testing.T) {
		for _, ci := range []string{parser.GitHub, parser.GitLab} {
			t.Run(ci, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{CI: &kickr.CI{Provider: ci}, Platform: ci},
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})

	t.Run("success_netlify", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Directory: "path/to/dir"}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Directory: "path/to/dir"}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{CI: &ci, Platform: ci.Provider},
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})

	t.Run("success_pages", func(t *testing.T) {
		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages, Directory: "path/to/dir"}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages, Directory: "path/to/dir"}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{CI: &ci, Platform: ci.Provider},
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
		// Arrange
		node := func(tc string) func(ctx context.Context, destdir string, config *types.KickrWrapper) error {
			return func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
				content := fmt.Appendf(nil, `{ "name": "kickr", "packageManager": "%s" }`+"\n", tc)
				return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), content, files.RwRR)
			}
		}

		for _, tc := range []string{"bun@1.1.6", "npm@7.0.0", "pnpm@9.0.0", "yarn@1.22.10"} {
			t.Run(tc, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI:       &kickr.CI{Provider: parser.GitHub},
						Platform: parser.GitHub,
					},
				}

				// Act & Assert
				test(ctx, t, config, node(tc))
			})
		}
	})

	t.Run("success_library", func(t *testing.T) {
		type testcase struct {
			Manager        string
			CI             string
			PackageManager string
		}

		// Arrange
		node := func(tc testcase) func(ctx context.Context, destdir string, config *types.KickrWrapper) error {
			return func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
				content := fmt.Appendf(nil, `{ "name": "kickr", "packageManager": "%s" }`+"\n", tc.PackageManager)
				return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), content, files.RwRR)
			}
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
						CI: &kickr.CI{
							Provider: tc.CI,
							Release:  &kickr.Release{Options: []string{kickr.OptionBackmerge}},
							Renovate: &kickr.Renovate{Auth: kickr.AuthPersonalToken},
						},
						Dependencies: &kickr.Dependencies{Manager: tc.Manager},
						Platform:     tc.CI,
					},
				}

				// Act & Assert
				test(ctx, t, config, node(tc))
			})
		}
	})

	t.Run("success_netlify", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Directory: "path/to/dir"}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingNetlify, Directory: "path/to/dir"}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
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
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_pages", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []kickr.CI{
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitHub, Website: &kickr.Website{Hosting: kickr.HostingPages, Directory: "path/to/dir"}},

			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages, Auto: true}},
			{Provider: parser.GitLab, Website: &kickr.Website{Hosting: kickr.HostingPages, Directory: "path/to/dir"}},
		}
		for _, ci := range cases {
			name := fmt.Sprint(ci.Provider, "_auto_", ci.Website.Auto)
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
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("success_helm", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

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

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})
}

func TestGenerate_MonoRepo(t *testing.T) {
	ctx := t.Context()

	type testcase struct {
		Provider string
		Hosting  string
	}

	golang := func(tc testcase) func(ctx context.Context, destdir string, config *types.KickrWrapper) error {
		return func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
			gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\ngo 1.23\n", tc.Provider)
			if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
			return nil
		}
	}

	hugo := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
		if err := os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx); err != nil {
			return fmt.Errorf("mkdir all: %w", err)
		}
		file, err := os.Create(filepath.Join(destdir, "docs", "hugo.toml"))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return file.Close()
	}

	node := func(_ context.Context, destdir string, _ *types.KickrWrapper) error {
		return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
			[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
	}

	t.Run("success_node_hugo_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.HostingPages},
			{Provider: parser.GitLab, Hosting: kickr.HostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.HostingPages},
			{Provider: parser.GitHub, Hosting: kickr.HostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
						},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, node, hugo)
			})
		}
	})

	t.Run("success_go_hugo_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.HostingPages},
			{Provider: parser.GitLab, Hosting: kickr.HostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.HostingPages},
			{Provider: parser.GitHub, Hosting: kickr.HostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
						},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(tc), hugo)
			})
		}
	})

	t.Run("success_go_node_doc", func(t *testing.T) {
		cases := []testcase{
			{Provider: parser.GitLab, Hosting: kickr.HostingPages},
			{Provider: parser.GitLab, Hosting: kickr.HostingNetlify},
			{Provider: parser.GitHub, Hosting: kickr.HostingPages},
			{Provider: parser.GitHub, Hosting: kickr.HostingNetlify},
		}
		for _, tc := range cases {
			name := fmt.Sprint(tc.Provider, "_", tc.Hosting)
			t.Run(name, func(t *testing.T) {
				// Arrange
				config := types.KickrWrapper{
					Kickr: kickr.Kickr{
						CI: &kickr.CI{
							Provider: tc.Provider,
							Website:  &kickr.Website{Hosting: tc.Hosting, Directory: "docs"},
						},
						Exclude:  []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit},
						Platform: tc.Provider,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(tc), node)
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
