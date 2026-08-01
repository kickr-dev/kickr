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

	directory string
	repo      *Repository

	Globs     map[string]any
	Languages map[string]any
}

var _ engine.Module = (*Module)(nil) // ensure interface is implemented

// Dir implements engine.Module.
func (m Module) Dir() string { return m.directory }

// Slug returns the directory slug representation.
func (m Module) Slug() string { return engine.ToSlug(m.directory) }

// Repository returns the repository owning this module.
//
// Repository-level configuration (Kickr, VCS) is only reachable through this accessor,
// so any edition has to go through the repository itself rather than the module.
func (m Module) Repository() *Repository { return m.repo }

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
	if m.repo.Docker == nil {
		return false
	}
	languages := []string{LanguageGo, LanguageHugo, LanguageNode}
	for _, language := range languages {
		if _, ok := m.Languages[language]; ok {
			return true
		}
	}
	return false
}

// HasMakefile returns truthy when the module should have a Makefile.
func (m Module) HasMakefile() bool {
	if slices.Contains(m.repo.Exclude, kickr.ExcludeMakefile) {
		return false
	}

	_, ok := m.Languages[LanguageNode]
	if ok && len(m.Languages) == 1 {
		return false
	}
	return true
}
