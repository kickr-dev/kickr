package initialize

import (
	huh "charm.land/huh/v2"
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// licenses holds the selectable license options, "none" first and mapped to
// an empty value so it clears config.License instead of literally storing "none".
var licenses = []huh.Option[string]{
	huh.NewOption("none", ""),
	huh.NewOption("agpl-3.0", "agpl-3.0"),
	huh.NewOption("apache-2.0", "apache-2.0"),
	huh.NewOption("bsd-2-clause", "bsd-2-clause"),
	huh.NewOption("bsd-3-clause", "bsd-3-clause"),
	huh.NewOption("bsl-1.0", "bsl-1.0"),
	huh.NewOption("cc0-1.0", "cc0-1.0"),
	huh.NewOption("epl-2.0", "epl-2.0"),
	huh.NewOption("gpl-2.0", "gpl-2.0"),
	huh.NewOption("gpl-3.0", "gpl-3.0"),
	huh.NewOption("lgpl-2.1", "lgpl-2.1"),
	huh.NewOption("mit", "mit"),
	huh.NewOption("mpl-2.0", "mpl-2.0"),
	huh.NewOption("unlicense", "unlicense"),
}

// License prompts the user to specify a license for the project.
func License(config *kickr.Kickr) *huh.Group {
	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("Would you like to specify a license ('none' to skip license) ?").
			Options(licenses...).
			Value(&config.License),
	)
}

var _ engine.FormGroup[kickr.Kickr] = License // ensure interface is implemented
