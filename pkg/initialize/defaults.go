package initialize

import (
	huh "charm.land/huh/v2"
	engine "github.com/kickr-dev/engine/pkg"

	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Defaults modifies the input kickr configuration to set default values (like version) during project initialization.
func Defaults(config *kickr.Kickr) *huh.Group {
	config.Version = 1
	return nil
}

var _ engine.FormGroup[kickr.Kickr] = Defaults // ensure interface is implemented
