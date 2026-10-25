package templates

import (
	"path"
	"slices"

	engine "gitlab.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/generate/types"
	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Chart returns the slice of templates related to helm chart generation.
//
// An existing Chart.yaml is kept as is, except for its kickr sections (description, workload dependency and maintainers), which are regenerated.
func Chart() (templates []engine.Template[types.Repository]) {
	application := path.Join("chart", "templates", "application.yaml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersChevron(),
		Globs:      []string{application + engine.TmplExtension},
		Out:        application,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
			return repo.Config.Helm == nil || repo.Config.IsHelmLibrary() || slices.Contains(repo.Config.Helm.Exclude, kickr.HelmExcludeWorkload)
		}, engine.PolicyDefault),
	})

	helmignore := path.Join("chart", ".helmignore")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{helmignore + engine.TmplExtension},
		Out:        helmignore,
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
			return repo.Config.Helm == nil
		}, engine.PolicyDefault),
	})

	chart := path.Join("chart", "Chart.yaml")
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters: engine.DelimitersBracket(),
		Globs:      []string{chart + engine.TmplExtension},
		Out:        chart,
		Sections: []engine.Section{
			{Begin: "# BEGIN_KICKR_DESCRIPTION", End: "# END_KICKR_DESCRIPTION"},
			{Begin: "# BEGIN_KICKR_WORKLOAD", End: "# END_KICKR_WORKLOAD"},
			{Begin: "# BEGIN_KICKR_MAINTAINERS", End: "# END_KICKR_MAINTAINERS"},
		},
		GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
			return repo.Config.Helm == nil
		}, engine.PolicyDefault),
	})

	valuesfiles := []string{
		path.Join("chart", "values.default.yaml"),
		path.Join("chart", "values.integration.yaml"),
		path.Join("chart", "values.production.yaml"),
		path.Join("chart", "values.review.yaml"),
		path.Join("chart", "values.staging.yaml"),
	}
	for _, src := range valuesfiles {
		templates = append(templates, engine.Template[types.Repository]{
			Delimiters: engine.DelimitersBracket(),
			Globs:      []string{src + engine.TmplExtension},
			Out:        src,
			GeneratePolicy: engine.RemoveIf(func(repo types.Repository) bool {
				return repo.Config.Helm == nil || repo.Config.IsHelmLibrary()
			}, engine.PolicyDefault),
		})
	}
	templates = append(templates, engine.Template[types.Repository]{
		Delimiters:  engine.DelimitersBracket(),
		EmptyPolicy: engine.PolicyKeepEmpty, // library and workload excluded charts get an empty values.yaml
		Globs:       []string{path.Join("chart", "values.yaml") + engine.TmplExtension},
		Out:         path.Join("chart", "values.yaml"),
		GeneratePolicy: func(out string, repo types.Repository) engine.GeneratePolicy {
			if repo.Config.Helm == nil {
				return engine.PolicyRemove
			}
			if repo.Config.IsHelmLibrary() || slices.Contains(repo.Config.Helm.Exclude, kickr.HelmExcludeWorkload) {
				return engine.SkipIfExists(out, engine.PolicyDefault)
			}
			return engine.PolicyDefault
		},
	})

	return templates
}
