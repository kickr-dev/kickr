package types //nolint:revive

import (
	"slices"

	engine "github.com/kickr-dev/engine/pkg"
	"github.com/kickr-dev/engine/pkg/parser"

	"github.com/kickr-dev/kickr/pkg/kickr/v1"
)

// Module represents one directory of a repository owning its own technology stack.
//
// Modules are enriched by parsers, each one adding the languages it detects in the directories it knows about.
// Kickr and VCS aren't known by parsers, they're the repository's ones, only reachable through Repository.
type Module struct {
	parser.Executables

	Config    kickr.Module
	Directory string
	Globs     map[string]any
	Languages map[string]any
	Parent    *Repository
}

var _ engine.Module = (*Module)(nil) // ensure interface is implemented

// Dir implements engine.Module.
func (m Module) Dir() string { return m.Directory }

// Slug returns the directory slug representation.
func (m Module) Slug() string { return engine.ToSlug(m.Directory) }

// SetLanguage sets a language with its specificities.
func (m *Module) SetLanguage(name string, value any) *Module {
	if m.Languages == nil {
		m.Languages = map[string]any{}
	}
	m.Languages[name] = value
	return m
}

// SetExecutables merges the input executables into the module's ones, per executable type.
func (m *Module) SetExecutables(executables parser.Executables) *Module {
	for name := range executables.Clis {
		m.AddCLI(name)
	}
	for name := range executables.Crons {
		m.AddCron(name)
	}
	for name := range executables.Jobs {
		m.AddJob(name)
	}
	for name := range executables.Workers {
		m.AddWorker(name)
	}
	return m
}

// SetGlob sets a glob by its name.
func (m *Module) SetGlob(name string, matches []string) *Module {
	if m.Globs == nil {
		m.Globs = map[string]any{}
	}
	m.Globs[name] = matches
	return m
}

// HasDocker returns truthy when the module should have a Dockerfile.
func (m Module) HasDocker() bool {
	if m.Parent.Config.Docker == nil || slices.Contains(m.Config.Exclude, kickr.ModuleExcludeDocker) {
		return false
	}
	if _, ok := m.Languages[LanguageGo]; ok {
		return m.Binaries() > 0
	}
	for _, language := range []string{LanguageHugo, LanguageNode} {
		if _, ok := m.Languages[language]; ok {
			return true
		}
	}
	return false
}

// HasMakefile returns truthy when the module should have a Makefile.
func (m Module) HasMakefile() bool {
	if slices.Contains(m.Parent.Config.Exclude, kickr.ExcludeMakefile) || slices.Contains(m.Config.Exclude, kickr.ModuleExcludeMakefile) {
		return false
	}
	if _, ok := m.Languages[LanguageNode]; ok && len(m.Languages) == 1 {
		return false
	}
	return true
}
