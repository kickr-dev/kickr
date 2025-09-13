package types

import (
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// KickrWrapper is the wrapper struct of kickr configuration file for a given project.
//
// This wrapper struct adds various properties, affected during kickr parsing to adds metadata around kickr configuration file.
//
// This includes automatic parsing of languages involved in the repository, various repository informations, etc.
type KickrWrapper struct {
	kickr.Kickr
	parser.Executables

	Globs     map[string]any
	Languages map[string]any
	VCS       parser.VCS
}

// SetLanguage sets a language with its specificities.
func (k *KickrWrapper) SetLanguage(name string, value any) {
	if k.Languages == nil {
		k.Languages = map[string]any{}
	}
	k.Languages[name] = value
}

// SetGlob sets a glob by its name.
func (k *KickrWrapper) SetGlob(name string, matches []string) {
	if k.Globs == nil {
		k.Globs = map[string]any{}
	}
	k.Globs[name] = matches
}

// Mono is a wrapper used for monorepository parsing.
// It helps identifying where is located a specific language.
type Mono[T any] struct {
	Directory string
	Specifics T
}
