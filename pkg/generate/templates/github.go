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
			return repo.GitHub == nil || !slices.Contains(repo.GitHub.Options, kickr.GitHubOptionsCodeQL)
		},
	})

	deployment := path.Join(".github", "workflows", "deployment.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(repo types.Repository) bool {
			if repo.GitHub == nil {
				return true
			}
			tf := repo.HasLanguage(types.LanguageTerraform) && repo.Terraform.Apply != ""                            //nolint:revive
			return repo.GitHub.Release == nil && !repo.HasDocker() && repo.Helm == nil && !tf && repo.Website == nil //nolint:revive
		},
	})

	integration := path.Join(".github", "workflows", "integration.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(integration),
		Out:        integration,
		Remove:     func(repo types.Repository) bool { return repo.GitHub == nil },
	})

	kickra := path.Join(".github", "workflows", "kickr.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{kickra + engine.TmplExtension},
		Out:        kickra,
		Remove: func(repo types.Repository) bool {
			return repo.GitHub == nil || !slices.ContainsFunc(repo.GitHub.Options, func(o string) bool {
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
			return repo.GitHub == nil || !slices.Contains(repo.GitHub.Options, kickr.GitHubOptionsLabeler)
		},
	})

	review := path.Join(".github", "workflows", "dependency-review.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{review + engine.TmplExtension},
		Out:        review,
		Remove:     func(repo types.Repository) bool { return repo.GitHub == nil },
	})

	scorecard := path.Join(".github", "workflows", "scorecard.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{scorecard + engine.TmplExtension},
		Out:        scorecard,
		Remove: func(repo types.Repository) bool {
			return repo.GitHub == nil || !slices.Contains(repo.GitHub.Options, kickr.GitHubOptionsOSSFScorecard)
		},
	})

	submission := path.Join(".github", "workflows", "dependency-submission.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{submission + engine.TmplExtension},
		Out:        submission,
		Remove: func(repo types.Repository) bool {
			return repo.GitHub == nil || !repo.HasLanguage(types.LanguageGo)
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
			return repo.GitHub == nil || !slices.Contains(repo.GitHub.Options, kickr.GitHubOptionsLabeler)
		},
	})

	release := path.Join(".github", "release.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{release + engine.TmplExtension},
		Out:        release,
		Remove:     func(repo types.Repository) bool { return repo.Platform != parser.GitHub },
	})

	return templates
}
