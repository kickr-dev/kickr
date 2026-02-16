package generate_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"dario.cat/mergo"
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

type testcase struct {
	kickr.Kickr

	Name string
}

func TestGenerate_NoLang(t *testing.T) {
	ctx := t.Context()

	t.Run("chart", func(t *testing.T) {
		cases := []testcase{
			{
				Name:  "github",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{}},
			},
			{
				Name:  "github_publish_auto",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			},
			{
				Name:  "github_publish_manual",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmPublishManual, Registry: "chartmuseum.example.com"}},
			},
			{
				Name:  "gitlab",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{}},
			},
			{
				Name:  "gitlab_publish_auto",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			},
			{
				Name:  "gitlab_publish_manual",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{Path: "path/to/kickr", Publish: kickr.HelmPublishManual, Registry: "chartmuseum.example.com"}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("kickr", func(t *testing.T) {
		cases := []testcase{
			{
				Name:  "github_github_app",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsKickrGitHubApp}}},
			},
			{
				Name:  "github_personal_token",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsKickrPersonalToken}}},
			},
			{
				Name:  "gitlab",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{Options: []string{kickr.GitLabOptionsKickr}}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("renovate", func(t *testing.T) {
		cases := []testcase{
			{
				Name:  "github_github_app",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsRenovateGitHubApp}}},
			},
			{
				Name:  "github_personal_token",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Options: []string{kickr.GitHubOptionsRenovatePersonalToken}}},
			},
			{
				Name:  "gitlab",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{Options: []string{kickr.GitLabOptionsRenovate}}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}

		t.Run("templates", func(t *testing.T) {
			// Arrange
			tmpl := func(_ context.Context, destdir string, _ *types.Repository) error {
				file, err := os.Create(filepath.Join(destdir, "template.tmpl"))
				if err != nil {
					return fmt.Errorf("create: %w", err)
				}
				return file.Close()
			}

			config := types.Repository{
				Kickr: kickr.Kickr{Exclude: []string{kickr.ExcludeMakefile}},
			}

			// Act & Assert
			test(ctx, t, config, tmpl)
		})
	})

	t.Run("precommit", func(t *testing.T) {
		cases := []testcase{
			{Name: "enabled_github", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "enabled_gitlab", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}}},
			{
				Name:  "disabled_github",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Exclude: []string{kickr.ExcludePreCommit}},
			},
			{
				Name:  "disabled_gitlab",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Exclude: []string{kickr.ExcludePreCommit}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitAutoCommit, kickr.PreCommitGitflowBranches, kickr.PreCommitConventionalCommits},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})

	t.Run("release", func(t *testing.T) {
		cases := []testcase{
			{
				Name:  "github",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Options: []string{kickr.ReleaseOptionsBackmerge}}}},
			},
			{
				Name:  "github_auto",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auto: true}}},
			},
			{
				Name:  "github_auth_github_app",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auth: kickr.ReleaseAuthGitHubApp}}},
			},
			{
				Name:  "github_auth_github_token",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auth: kickr.ReleaseAuthGitHubToken}}},
			},
			{
				Name:  "github_auth_personal_token",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{Auth: kickr.ReleaseAuthPersonalToken}}},
			},
			{
				Name:  "gitlab",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{Options: []string{kickr.ReleaseOptionsBackmerge}}}},
			},
			{
				Name:  "gitlab_auto",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{Auto: true}}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})
}

func TestGenerate_Shell(t *testing.T) {
	ctx := t.Context()

	shell := func(_ context.Context, destdir string, _ *types.Repository) error {
		return os.WriteFile(filepath.Join(destdir, "script.sh"), []byte("#!/bin/sh\n"), files.RwxRxRxRx)
	}

	cases := []testcase{
		{Name: "github", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}}},
		{Name: "gitlab", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}}},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			config := types.Repository{
				Kickr: merge(t, kickr.Kickr{
					Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
				}, tc.Kickr),
			}

			// Act & Assert
			test(ctx, t, config, shell)
		})
	}

	t.Run("precommit", func(t *testing.T) {
		cases := []testcase{
			{Name: "disabled", Kickr: kickr.Kickr{Exclude: []string{kickr.ExcludePreCommit}}},
			{Name: "enabled", Kickr: kickr.Kickr{}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, shell)
			})
		}
	})
}

func TestGenerate_Golang(t *testing.T) {
	ctx := t.Context()

	t.Run("cli", func(t *testing.T) {
		// Arrange
		golang := func(provider string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", provider)
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

		cases := []testcase{
			{Name: "github", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{}}}},
			{Name: "gitlab", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{}}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{Kickr: merge(t, kickr.Kickr{}, tc.Kickr)}

				// Act & Assert
				test(ctx, t, config, golang(tc.Name))
			})
		}
	})

	t.Run("library", func(t *testing.T) {
		// Arrange
		golang := func(platform string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", platform)
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}
				return nil
			}
		}

		for _, platform := range []string{parser.GitLab, parser.GitHub} {
			t.Run(platform, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitGomodTidy},
						Platform:  platform,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(platform))
			})
		}
	})

	t.Run("multiple_bin_helm", func(t *testing.T) {
		// Arrange
		golang := func(provider string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", provider)
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

		cases := []testcase{
			{
				Name: "github",
				Kickr: kickr.Kickr{
					GitHub: &kickr.GitHub{
						Options: []string{
							kickr.GitHubOptionsCodecov,
							kickr.GitHubOptionsCodeQL,
							kickr.GitHubOptionsHardenRunner,
							kickr.GitHubOptionsLabeler,
							kickr.GitHubOptionsOSSFScorecard,
							kickr.GitHubOptionsSonarQube,
							kickr.GitHubOptionsStepSecurityActions,
						},
						Release: &kickr.Release{},
					},
				},
			},
			{
				Name: "gitlab",
				Kickr: kickr.Kickr{
					GitLab: &kickr.GitLab{
						Options: []string{kickr.GitLabOptionsSonarQube},
						Release: &kickr.Release{},
					},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Docker: &kickr.Docker{Path: "path/to/registry", Registry: "registry.example.com"},
						Helm: &kickr.Helm{
							Deploy:       kickr.HelmDeployManual,
							Environments: []string{kickr.EnvironmentStaging, kickr.EnvironmentProduction},
							Path:         "path/to/repository",
							Publish:      kickr.HelmPublishManual,
							Registry:     "chartmuseum.example.com",
						},
						Description: "A useful project description",
						Exclude:     []string{kickr.ExcludeRenovate, kickr.ExcludeShell},
						PreCommit:   []string{kickr.PreCommitGolangciLint},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, golang(tc.Name))
			})
		}
	})

	t.Run("multiple_libraries", func(t *testing.T) {
		// Arrange
		golang := func(platform string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				if err := os.WriteFile(filepath.Join(destdir, parser.FileGowork), []byte("go 1.23\n\nuse (\n\t./kickr\n\t./engine\n)\n"), files.RwRR); err != nil {
					return fmt.Errorf("write file: %w", err)
				}

				for _, dir := range []string{"kickr", "engine"} {
					if err := os.MkdirAll(filepath.Join(destdir, dir), files.RwxRxRxRx); err != nil {
						return fmt.Errorf("mkdir all: %w", err)
					}
					gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/%s\n\ngo 1.23\n", platform, dir)
					if err := os.WriteFile(filepath.Join(destdir, dir, parser.FileGomod), gomod, files.RwRR); err != nil {
						return fmt.Errorf("write file: %w", err)
					}
				}
				return nil
			}
		}

		for _, platform := range []string{parser.GitLab, parser.GitHub} {
			t.Run(platform, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						PreCommit: []string{kickr.PreCommitGomodTidy},
						Platform:  platform,
					},
				}

				// Act & Assert
				test(ctx, t, config, golang(platform))
			})
		}
	})
}

func TestGenerate_Hugo(t *testing.T) {
	ctx := t.Context()

	hugo := func(_ context.Context, destdir string, _ *types.Repository) error {
		file, err := os.Create(filepath.Join(destdir, "hugo.toml"))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return file.Close()
	}

	t.Run("no_website", func(t *testing.T) {
		cases := []testcase{
			{Name: "github", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{Kickr: merge(t, kickr.Kickr{}, tc.Kickr)}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})

	t.Run("hosting", func(t *testing.T) {
		cases := []testcase{
			{
				Name:  "github_netlify_auto",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			},
			{
				Name:  "github_netlify",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},
			},
			{
				Name:  "gitlab_netlify_auto",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			},
			{
				Name:  "gitlab_netlify",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},
			},

			{
				Name:  "github_pages_auto",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			},
			{
				Name:  "github_pages",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},
			},
			{
				Name:  "gitlab_pages_auto",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			},
			{
				Name:  "gitlab_pages",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, hugo)
			})
		}
	})
}

func TestGenerate_Node(t *testing.T) {
	ctx := t.Context()

	t.Run("package_managers", func(t *testing.T) {
		// Arrange
		node := func(tc string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				content := fmt.Appendf(nil, `{ "name": "kickr", "packageManager": "%s" }`+"\n", tc)
				return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON), content, files.RwRR)
			}
		}

		for _, tc := range []string{"bun@1.1.6", "npm@7.0.0", "pnpm@9.0.0", "yarn@1.22.10"} {
			t.Run(tc, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: kickr.Kickr{
						GitHub:   &kickr.GitHub{},
						Platform: parser.GitHub,
					},
				}

				// Act & Assert
				test(ctx, t, config, node(tc))
			})
		}
	})

	t.Run("library", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Name: "github_bun", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{}}}},
			{Name: "gitlab_bun", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{}}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{Kickr: merge(t, kickr.Kickr{}, tc.Kickr)}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("hosting", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{
				Name:  "github_netlify_auto",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			},
			{
				Name:  "github_netlify",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},
			},
			{
				Name:  "gitlab_netlify_auto",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify, Auto: true}},
			},
			{
				Name:  "gitlab_netlify",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingNetlify}},
			},

			{
				Name:  "github_pages_auto",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			},
			{
				Name:  "github_pages",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},
			},
			{
				Name:  "gitlab_pages_auto",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages, Auto: true}},
			},
			{
				Name:  "gitlab_pages",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Website: &kickr.Website{Hosting: kickr.WebsiteHostingPages}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
						Website: tc.Website,
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})

	t.Run("helm", func(t *testing.T) {
		// Arrange
		node := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js" }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{
				Name:  "github",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{}},
			},
			{
				Name: "github_deploy_auto",
				Kickr: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentReview}},
				},
			},
			{
				Name: "github_deploy_manual",
				Kickr: kickr.Kickr{
					GitHub: &kickr.GitHub{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployManual, Environments: []string{kickr.EnvironmentIntegration}},
				},
			},
			{
				Name:  "github_publish_auto",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			},
			{
				Name:  "github_publish_manual",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishManual}},
			},

			{
				Name:  "gitlab",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{}},
			},
			{
				Name: "gitlab_deploy_auto",
				Kickr: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployAuto, Environments: []string{kickr.EnvironmentReview}},
				},
			},
			{
				Name: "gitlab_deploy_manual",
				Kickr: kickr.Kickr{
					GitLab: &kickr.GitLab{},
					Helm:   &kickr.Helm{Deploy: kickr.HelmDeployManual, Environments: []string{kickr.EnvironmentIntegration}},
				},
			},
			{
				Name:  "gitlab_publish_auto",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishAuto}},
			},
			{
				Name:  "gitlab_publish_manual",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Helm: &kickr.Helm{Publish: kickr.HelmPublishManual}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeRenovate},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, node)
			})
		}
	})
}

func TestGenerate_Terraform(t *testing.T) {
	ctx := t.Context()

	t.Run("multiple_modules", func(t *testing.T) {
		// Arrange
		terraform := func(subdir string) func(ctx context.Context, destdir string, config *types.Repository) error {
			return func(_ context.Context, destdir string, _ *types.Repository) error {
				if err := os.MkdirAll(filepath.Join(destdir, subdir), files.RwxRxRxRx); err != nil {
					return fmt.Errorf("mkdir all: %w", err)
				}
				return os.WriteFile(filepath.Join(destdir, subdir, "main.tf"), []byte(
					`terraform { backend "http" {} }`+"\n"+
						`variable "my_secret" { sensitive = true }`+"\n"+
						`variable "github_var" {}`+"\n"+
						`variable "my_var" {}`+"\n"), files.RwRR)
			}
		}

		cases := []testcase{
			{
				Name: "github_opentofu_apply_manual",
				Kickr: kickr.Kickr{
					GitHub:    &kickr.GitHub{Release: &kickr.Release{}},
					Terraform: &kickr.Terraform{Apply: kickr.TerraformApplyManual, Engine: kickr.TerraformEngineOpenTofu},
				},
			},
			{
				Name: "github_terraform_apply_auto",
				Kickr: kickr.Kickr{
					GitHub:    &kickr.GitHub{Release: &kickr.Release{}},
					Terraform: &kickr.Terraform{Apply: kickr.TerraformApplyAuto, Engine: kickr.TerraformEngineTerraform},
				},
			},

			{
				Name: "gitlab_opentofu_apply_manual",
				Kickr: kickr.Kickr{
					GitLab:    &kickr.GitLab{Release: &kickr.Release{}},
					Terraform: &kickr.Terraform{Apply: kickr.TerraformApplyManual, Engine: kickr.TerraformEngineOpenTofu},
				},
			},
			{
				Name: "gitlab_terraform_apply_auto",
				Kickr: kickr.Kickr{
					GitLab:    &kickr.GitLab{Release: &kickr.Release{}},
					Terraform: &kickr.Terraform{Apply: kickr.TerraformApplyAuto, Engine: kickr.TerraformEngineTerraform},
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Terraform: &kickr.Terraform{
							Environments: []string{kickr.EnvironmentProduction},
							Modules:      []string{"modules/one", "modules/two"},
						},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, terraform(filepath.Join("modules", "one")), terraform(filepath.Join("modules", "two")))
			})
		}
	})

	t.Run("root_module", func(t *testing.T) {
		// Arrange
		terraform := func(_ context.Context, destdir string, _ *types.Repository) error {
			return os.WriteFile(filepath.Join(destdir, "main.tf"), []byte(`terraform { backend "s3" {} }`+"\n"), files.RwRR)
		}

		cases := []testcase{
			{Name: "github", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}}},
			{Name: "gitlab", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						PreCommit: []string{kickr.PreCommitTerraform},
						Terraform: &kickr.Terraform{Environments: []string{kickr.EnvironmentProduction}},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, terraform)
			})
		}
	})
}

func TestGenerate_MonoRepo(t *testing.T) {
	ctx := t.Context()

	golang := func(provider string) func(ctx context.Context, destdir string, config *types.Repository) error {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			gomod := fmt.Appendf(nil, "module %s.com/kickr-dev/kickr\n\ngo 1.23\n", provider)
			if err := os.WriteFile(filepath.Join(destdir, parser.FileGomod), gomod, files.RwRR); err != nil {
				return fmt.Errorf("write file: %w", err)
			}
			return nil
		}
	}

	hugo := func(_ context.Context, destdir string, _ *types.Repository) error {
		if err := os.MkdirAll(filepath.Join(destdir, "docs"), files.RwxRxRxRx); err != nil {
			return fmt.Errorf("mkdir all: %w", err)
		}
		file, err := os.Create(filepath.Join(destdir, "docs", "hugo.toml"))
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return file.Close()
	}

	node := func(subdir string) engine.Parser[types.Repository] {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, subdir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			return os.WriteFile(filepath.Join(destdir, subdir, parser.FilePackageJSON),
				[]byte(`{ "name": "kickr", "packageManager": "bun@1.1.6", "main": "index.js", "private": true }`+"\n"), files.RwRR)
		}
	}

	terraform := func(subdir string) func(ctx context.Context, destdir string, config *types.Repository) error {
		return func(_ context.Context, destdir string, _ *types.Repository) error {
			if err := os.MkdirAll(filepath.Join(destdir, subdir), files.RwxRxRxRx); err != nil {
				return fmt.Errorf("mkdir all: %w", err)
			}
			return os.WriteFile(filepath.Join(destdir, subdir, "main.tf"), []byte(`variable "my_var" {}`+"\n"), files.RwRR)
		}
	}

	cases := []testcase{
		{
			Name: "github_netlify",
			Kickr: kickr.Kickr{
				GitHub:   &kickr.GitHub{},
				Website:  &kickr.Website{Hosting: kickr.WebsiteHostingNetlify},
				Platform: parser.GitHub,
			},
		},
		{
			Name: "github_pages",
			Kickr: kickr.Kickr{
				GitHub:   &kickr.GitHub{},
				Website:  &kickr.Website{Hosting: kickr.WebsiteHostingPages},
				Platform: parser.GitHub,
			},
		},
		{
			Name: "gitlab_netlify",
			Kickr: kickr.Kickr{
				GitLab:   &kickr.GitLab{},
				Website:  &kickr.Website{Hosting: kickr.WebsiteHostingNetlify},
				Platform: parser.GitLab,
			},
		},
		{
			Name: "gitlab_pages",
			Kickr: kickr.Kickr{
				GitLab:   &kickr.GitLab{},
				Website:  &kickr.Website{Hosting: kickr.WebsiteHostingPages},
				Platform: parser.GitLab,
			},
		},
	}

	t.Run("node_hugo_doc", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Website: &kickr.Website{Directory: "docs"},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, node(""), hugo)
			})
		}
	})

	t.Run("go_hugo_doc", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Website: &kickr.Website{Directory: "docs"},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, golang(tc.Platform), hugo)
			})
		}
	})

	t.Run("go_node_doc", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit, kickr.ExcludeRenovate},
						Website: &kickr.Website{Directory: "docs"},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, golang(tc.Platform), node("docs"))
			})
		}
	})

	t.Run("go_self_terraform", func(t *testing.T) {
		cases := []testcase{
			{Name: "github", Kickr: kickr.Kickr{GitHub: &kickr.GitHub{}, Platform: parser.GitHub}},
			{Name: "gitlab", Kickr: kickr.Kickr{GitLab: &kickr.GitLab{}, Platform: parser.GitLab}},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						Exclude:   []string{kickr.ExcludeMakefile, kickr.ExcludePreCommit},
						Terraform: &kickr.Terraform{Engine: kickr.TerraformEngineOpenTofu, Modules: []string{".terraform"}},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config, golang(tc.Platform), terraform(".terraform"))
			})
		}
	})
}

func TestGenerate_MultiPlatforms(t *testing.T) {
	ctx := t.Context()

	t.Run("multiple_ci", func(t *testing.T) {
		cases := []testcase{
			{
				Name:  "github_release",
				Kickr: kickr.Kickr{GitHub: &kickr.GitHub{Release: &kickr.Release{}}},
			},
			{
				Name:  "gitlab_release",
				Kickr: kickr.Kickr{GitLab: &kickr.GitLab{Release: &kickr.Release{}}},
			},
		}
		for _, tc := range cases {
			t.Run(tc.Name, func(t *testing.T) {
				// Arrange
				config := types.Repository{
					Kickr: merge(t, kickr.Kickr{
						GitHub:  &kickr.GitHub{},
						GitLab:  &kickr.GitLab{},
						Exclude: []string{kickr.ExcludeMakefile, kickr.ExcludeShell},
						PreCommit: []string{
							kickr.PreCommitAutoCommit,
							kickr.PreCommitConventionalCommits,
							kickr.PreCommitGitflowBranches,
							kickr.PreCommitGolangciLint,
							kickr.PreCommitGomodTidy,
							kickr.PreCommitTerraform,
						},
					}, tc.Kickr),
				}

				// Act & Assert
				test(ctx, t, config)
			})
		}
	})
}

func ParserInfo(_ context.Context, _ string, config *types.Repository) error {
	config.VCS = parser.VCS{
		Platform:    config.Platform,
		ProjectHost: config.Platform + ".com",
		ProjectName: "kickr",
		ProjectPath: "kickr-dev/kickr",
	}
	return nil
}

// merge combines the two input kickr configurations.
//
// Should only be used to avoid conditions between CICD provides
// and not as a general use case.
func merge(t testing.TB, base, complement kickr.Kickr) kickr.Kickr {
	t.Helper()

	require.NoError(t, mergo.Merge(&base, complement, mergo.WithAppendSlice))
	if base.Platform == "" {
		if base.GitHub != nil {
			base.Platform = parser.GitHub
		}
		if base.GitLab != nil {
			base.Platform = parser.GitLab
		}
	}
	return base
}

// test verifies every generation with provided config, parser and t.Name folder expected results.
func test(ctx context.Context, t *testing.T, config types.Repository, parsers ...engine.Parser[types.Repository]) {
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
		slices.Concat(parsers, []engine.Parser[types.Repository]{
			// must be kept first since it parses Git informations (useful for next parsers)
			// generate.ParserGit,
			ParserInfo,

			generate.ParserGlob,
			generate.ParserGolang,
			generate.ParserNode,
			generate.ParserTerraform,

			// must be kept last since it marshals config and merges it with chart overrides
			generate.ParserHelm,
		}),
		[]engine.Generator[types.Repository]{
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.CodeCov(), templates.Sonar())),                              // coverage
			engine.GeneratorTemplates(templates.FS(), slices.Concat(templates.GitHub(), templates.GitLab(), templates.SemanticRelease())), // ci
			engine.GeneratorTemplates(templates.FS(), templates.Chart()),                                                                  // chart
			engine.GeneratorTemplates(templates.FS(), templates.Docker()),                                                                 // docker
			engine.GeneratorTemplates(templates.FS(), templates.Golang()),                                                                 // golang
			engine.GeneratorTemplates(templates.FS(), templates.Makefile()),                                                               // makefile
			engine.GeneratorTemplates(templates.FS(), templates.Misc()),                                                                   // misc
			engine.GeneratorTemplates(templates.FS(), templates.Renovate()),                                                               // renovate
			engine.GeneratorTemplates(templates.FS(), templates.Terraform()),                                                              // terraform
		})

	// Assert
	require.NoError(t, err)
	assert.NoError(t, compare.Dirs(assertdir, destdir))
}
