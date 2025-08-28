package templates

import (
	"path"

	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
)

// Chart returns the slice of templates related to helm chart generation.
func Chart() []engine.Template[types.KickrGen] {
	var templates []engine.Template[types.KickrGen] //nolint:prealloc

	tmplfiles := []string{
		path.Join("chart", "templates", "_helpers.tpl"),
		path.Join("chart", "templates", "configmap.yaml"),
		path.Join("chart", "templates", "cronjob.yaml"),
		path.Join("chart", "templates", "deployment.yaml"),
		path.Join("chart", "templates", "hpa.yaml"),
		path.Join("chart", "templates", "job.yaml"),
		path.Join("chart", "templates", "service.yaml"),
		path.Join("chart", "templates", "serviceaccount.yaml"),
	}
	for _, src := range tmplfiles {
		templates = append(templates, engine.Template[types.KickrGen]{
			Delimiters: engine.DelimitersChevron(),
			Globs:      []string{src + engine.TmplExtension},
			Out:        src,
			Remove: func(config types.KickrGen) bool {
				return config.CI == nil || config.CI.Helm == nil
			},
		})
	}

	chartfiles := []string{
		path.Join("chart", ".kickr"),
		path.Join("chart", ".helmignore"),
		path.Join("chart", "Chart.yaml"),
		path.Join("chart", "charts", ".gitkeep"),
	}
	for _, src := range chartfiles {
		templates = append(templates, engine.Template[types.KickrGen]{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{src + engine.TmplExtension},
			Out:        src,
			Remove: func(config types.KickrGen) bool {
				return config.CI == nil || config.CI.Helm == nil
			},
		})
	}

	templates = append(templates, engine.Template[types.KickrGen]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      engine.GlobsWithPart(path.Join("chart", "values.yaml")),
		Out:        path.Join("chart", "values.yaml"),
		Remove: func(config types.KickrGen) bool {
			return config.CI == nil || config.CI.Helm == nil
		},
	})

	return templates
}
