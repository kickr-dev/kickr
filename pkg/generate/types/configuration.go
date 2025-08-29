package types

import (
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

type KickrGen struct {
	kickr.Kickr
	parser.Executables

	Globs     map[string]any
	Languages map[string]any
	VCS       parser.VCS
}

// SetLanguage sets a language with its specificities.
func (k *KickrGen) SetLanguage(name string, value any) {
	if k.Languages == nil {
		k.Languages = map[string]any{}
	}
	k.Languages[name] = value
}

// SetGlob sets a glob by its name.
func (k *KickrGen) SetGlob(name string) {
	if k.Globs == nil {
		k.Globs = map[string]any{}
	}
	k.Globs[name] = struct{}{}
}
