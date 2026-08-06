package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GitHub returns the slice of templates related to GitHub configuration.
func GitHub() []engine.Template[types.Repository] {
	return slices.Concat(githubWorkflow(), githubConfig())
}

func githubWorkflow() (templates []engine.Template[types.Repository]) { //nolint:funlen
	codeql := path.Join(".github", "workflows", "codeql.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{codeql + engine.TmplExtension},
		Out:        codeql,
		Remove: func(repo types.Repository) bool {
			return repo.Config.GitHub == nil || !slices.Contains(repo.Config.GitHub.Options, kickr.GitHubOptionsCodeQL)
		},
	})

	deployment := path.Join(".github", "workflows", "deployment.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(repo types.Repository) bool {
			if repo.Config.GitHub == nil {
				return true
			}
			hasApply := slices.ContainsFunc(repo.Config.Modules, func(m kickr.Module) bool {
				return m.Terraform != nil && m.Terraform.Apply != ""
			})
			hasDeployment := len(repo.ModulesWithDeployment(kickr.DeploymentTargetNetlify)) > 0 || len(repo.ModulesWithDeployment(kickr.DeploymentTargetPages)) > 0
			return repo.Config.GitHub.Release == nil && !repo.HasDocker() && repo.Config.Helm == nil && !hasApply && !hasDeployment //nolint:revive
		},
	})

	integration := path.Join(".github", "workflows", "integration.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(integration),
		Out:        integration,
		Remove:     func(repo types.Repository) bool { return repo.Config.GitHub == nil },
	})

	kickra := path.Join(".github", "workflows", "kickr.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{kickra + engine.TmplExtension},
		Out:        kickra,
		Remove: func(repo types.Repository) bool {
			return repo.Config.GitHub == nil || !slices.ContainsFunc(repo.Config.GitHub.Options, func(o string) bool {
				return o == kickr.GitHubOptionsKickrGitHubApp || o == kickr.GitHubOptionsKickrPersonalToken
			})
		},
	})

	labeler := path.Join(".github", "workflows", "labeler.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{labeler + engine.TmplExtension},
		Out:        labeler,
		Remove: func(repo types.Repository) bool {
			return repo.Config.GitHub == nil || !slices.Contains(repo.Config.GitHub.Options, kickr.GitHubOptionsLabeler)
		},
	})

	review := path.Join(".github", "workflows", "dependency-review.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{review + engine.TmplExtension},
		Out:        review,
		Remove:     func(repo types.Repository) bool { return repo.Config.GitHub == nil },
	})

	scorecard := path.Join(".github", "workflows", "scorecard.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{scorecard + engine.TmplExtension},
		Out:        scorecard,
		Remove: func(repo types.Repository) bool {
			return repo.Config.GitHub == nil || !slices.Contains(repo.Config.GitHub.Options, kickr.GitHubOptionsOSSFScorecard)
		},
	})

	submission := path.Join(".github", "workflows", "dependency-submission.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{submission + engine.TmplExtension},
		Out:        submission,
		Remove: func(repo types.Repository) bool {
			return repo.Config.GitHub == nil || len(repo.ModulesWith(types.LanguageGo)) == 0
		},
	})

	return templates
}

func githubConfig() (templates []engine.Template[types.Repository]) {
	labeler := path.Join(".github", "labeler.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{labeler + engine.TmplExtension},
		Out:        labeler,
		Remove: func(repo types.Repository) bool {
			return repo.Config.GitHub == nil || !slices.Contains(repo.Config.GitHub.Options, kickr.GitHubOptionsLabeler)
		},
	})

	release := path.Join(".github", "release.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{release + engine.TmplExtension},
		Out:        release,
		Remove:     func(repo types.Repository) bool { return repo.VCS.Platform != parser.GitHub },
	})

	return templates
}
