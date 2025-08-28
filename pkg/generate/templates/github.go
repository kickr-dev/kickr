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
func GitHub() []engine.Template[types.KickrGen] {
	return slices.Concat(githubWorkflow(), githubConfig())
}

func githubWorkflow() []engine.Template[types.KickrGen] {
	var templates []engine.Template[types.KickrGen]

	codeql := path.Join(".github", "workflows", "codeql.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{codeql + engine.TmplExtension},
		Out:        codeql,
		Remove: func(config types.KickrGen) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionCodeQL)
		},
	})

	deployment := path.Join(".github", "workflows", "deployment.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(config types.KickrGen) bool {
			return !config.IsCI(parser.GitHub) || (!config.HasHelmPublish() && !config.HasDeployment() && !config.HasRelease())
		},
	})

	integration := path.Join(".github", "workflows", "integration.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(integration),
		Out:        integration,
		Remove:     func(config types.KickrGen) bool { return !config.IsCI(parser.GitHub) },
	})

	labeler := path.Join(".github", "workflows", "labeler.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{labeler + engine.TmplExtension},
		Out:        labeler,
		Remove: func(config types.KickrGen) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionLabeler)
		},
	})

	review := path.Join(".github", "workflows", "dependency-review.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{review + engine.TmplExtension},
		Out:        review,
		Remove:     func(config types.KickrGen) bool { return !config.IsCI(parser.GitHub) },
	})

	scorecard := path.Join(".github", "workflows", "scorecard.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{scorecard + engine.TmplExtension},
		Out:        scorecard,
		Remove: func(config types.KickrGen) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionScoreCardOSSF)
		},
	})

	submission := path.Join(".github", "workflows", "dependency-submission.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{submission + engine.TmplExtension},
		Out:        submission,
		Remove: func(config types.KickrGen) bool {
			_, ok := config.Languages["go"]
			return !ok || !config.IsCI(parser.GitHub)
		},
	})

	return templates
}

func githubConfig() []engine.Template[types.KickrGen] {
	var templates []engine.Template[types.KickrGen]

	labeler := path.Join(".github", "labeler.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{labeler + engine.TmplExtension},
		Out:        labeler,
		Remove: func(config types.KickrGen) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.OptionLabeler)
		},
	})

	release := path.Join(".github", "release.yml")
	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{release + engine.TmplExtension},
		Out:        release,
		Remove:     func(config types.KickrGen) bool { return config.Platform != parser.GitHub },
	})

	return templates
}
