package templates

import (
	"path"
	"slices"
	"strings"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	kickr "github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// GitHub returns the slice of templates related to GitHub configuration.
func GitHub() []engine.Template[types.Repository] {
	return slices.Concat(githubWorkflow(), githubConfig())
}

func githubWorkflow() (templates []engine.Template[types.Repository]) {
	codeql := path.Join(".github", "workflows", "codeql.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{codeql + engine.TmplExtension},
		Out:        codeql,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionsCodeQL)
		},
	})

	deployment := path.Join(".github", "workflows", "deployment.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitHub) || //nolint:revive
				(config.CI.Release == nil && config.Docker == nil && config.Helm == nil && config.Website == nil &&
					(config.Terraform == nil || config.Terraform.Apply == ""))
		},
	})

	integration := path.Join(".github", "workflows", "integration.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(integration),
		Out:        integration,
		Remove:     func(config types.Repository) bool { return !config.IsCI(parser.GitHub) },
	})

	kickra := path.Join(".github", "workflows", "kickr.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{kickra + engine.TmplExtension},
		Out:        kickra,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitHub) || !slices.ContainsFunc(config.CI.Options, func(v string) bool { return strings.HasPrefix(v, "kickr:") })
		},
	})

	labeler := path.Join(".github", "workflows", "labeler.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{labeler + engine.TmplExtension},
		Out:        labeler,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionsLabeler)
		},
	})

	review := path.Join(".github", "workflows", "dependency-review.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{review + engine.TmplExtension},
		Out:        review,
		Remove:     func(config types.Repository) bool { return !config.IsCI(parser.GitHub) },
	})

	scorecard := path.Join(".github", "workflows", "scorecard.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{scorecard + engine.TmplExtension},
		Out:        scorecard,
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionsOSSFScorecard)
		},
	})

	submission := path.Join(".github", "workflows", "dependency-submission.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{submission + engine.TmplExtension},
		Out:        submission,
		Remove: func(config types.Repository) bool {
			_, ok := config.Languages["go"]
			return !ok || !config.IsCI(parser.GitHub)
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
		Remove: func(config types.Repository) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionsLabeler)
		},
	})

	release := path.Join(".github", "release.yml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{release + engine.TmplExtension},
		Out:        release,
		Remove:     func(config types.Repository) bool { return config.Platform != parser.GitHub },
	})

	return templates
}
