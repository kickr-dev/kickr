package templates

import (
	"path"
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	kickr "github.com/kickr-dev/kickr/pkg/configuration"
)

// GitHub returns the slice of templates related to GitHub configuration.
func GitHub() []engine.Template[kickr.Config] {
	return slices.Concat(githubWorkflow(), githubConfig())
}

func githubWorkflow() []engine.Template[kickr.Config] {
	var templates []engine.Template[kickr.Config]

	integration := path.Join(".github", "workflows", "integration.yml")
	templates = append(templates, engine.Template[kickr.Config]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(integration),
		Out:        integration,
		Remove:     func(config kickr.Config) bool { return !config.IsCI(parser.GitHub) },
	})

	deployment := path.Join(".github", "workflows", "deployment.yml")
	templates = append(templates, engine.Template[kickr.Config]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      engine.GlobsWithPart(deployment),
		Out:        deployment,
		Remove: func(config kickr.Config) bool {
			return !config.IsCI(parser.GitHub) || (!config.HasHelmPublish() && !config.HasDeployment() && !config.HasRelease())
		},
	})

	codeql := path.Join(".github", "workflows", "codeql.yml")
	templates = append(templates, engine.Template[kickr.Config]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{codeql + engine.TmplExtension},
		Out:        codeql,
		Remove: func(config kickr.Config) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.CodeQL)
		},
	})

	dependencies := path.Join(".github", "workflows", "dependencies.yml")
	templates = append(templates, engine.Template[kickr.Config]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{dependencies + engine.TmplExtension},
		Out:        dependencies,
		Remove: func(config kickr.Config) bool {
			_, ok := config.Languages["go"]
			return !ok || !config.IsCI(parser.GitHub)
		},
	})

	labeler := path.Join(".github", "workflows", "labeler.yml")
	templates = append(templates, engine.Template[kickr.Config]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{labeler + engine.TmplExtension},
		Out:        labeler,
		Remove: func(config kickr.Config) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.Labeler)
		},
	})

	return templates
}

func githubConfig() []engine.Template[kickr.Config] {
	var templates []engine.Template[kickr.Config]

	labeler := path.Join(".github", "labeler.yml")
	templates = append(templates, engine.Template[kickr.Config]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{labeler + engine.TmplExtension},
		Out:        labeler,
		Remove: func(config kickr.Config) bool {
			return !config.IsCI(parser.GitHub) || !slices.Contains(config.CI.Options, kickr.Labeler)
		},
	})

	release := path.Join(".github", "release.yml")
	templates = append(templates, engine.Template[kickr.Config]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{release + engine.TmplExtension},
		Out:        release,
		Remove:     func(config kickr.Config) bool { return config.Platform != parser.GitHub },
	})

	return templates
}
