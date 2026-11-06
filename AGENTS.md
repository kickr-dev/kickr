# kickr

kickr generates and maintains a repository layout (CI/CD, Helm chart, Docker, linters, community files) from `.kickr.yml` and what its parsers detect.
It builds on the `gitlab.com/kickr-dev/engine` SDK for parsers, generators and template rendering.

The [CONTRIBUTING.md](CONTRIBUTING.md) lists the commit types and the release each one triggers.
Read it before writing a commit, since `refactor` releases a minor version here.

## Where things go

- The configuration schema lives in `.schemas/kickr.v1.schema.yml`, see [Schema](#schema).
- A predicate reading only the configuration goes in `pkg/kickr/v1/methods.go`.
- A predicate reading parsed state (languages, globs, executables) goes on `Repository` or `Module` in `pkg/generate/types`.
- A parser goes in `pkg/generate/parser_<name>.go`, a generator in `pkg/generate/generator_<name>.go`.
- A template set goes in `pkg/generate/templates/<topic>.go`, its files under `pkg/generate/templates/_templates/`.
- A standalone command (*e.g.* `kickr chart`) goes in `internal/cobra/<name>.go` with `GroupID: groupStandalone`.
- Shared test helpers go in `testutils/`.

## Schema

The schema `.schemas/kickr.v1.schema.yml` is the source of truth.
Never edit `.schemas/kickr.v1.schema.json`, `pkg/kickr/v1/kickr.go` or `pkg/kickr/v1/constants.go` by hand.

1. Edit the yml.
2. Regenerate the JSON with `cd ./.schemas && go generate ./...`.
3. Update `kickr.go` (structs, GoDoc) and `constants.go` (enums) with the `schema-to-go` skill.

Prefer an entry in an existing `options` or `exclude` enum (`group:item`, *e.g.* `vendor:workload`) over a new scalar key.

## Go code

- Never add a language-specific method (*e.g.* `HasHugo()`) to `Repository` or `Module`.
- Put language logic in a template or a dedicated parser.
- Declare a lookup list in the function that needs it, never as a package-level `var`.
- Look for a helper in `pkg/generate/types` and `pkg/kickr/v1/methods.go` before writing a comparison.
- Common helpers are `Module.HasDocker()`, `Module.HasBuildTool()`, `Repository.ModulesWith()` and `Repository.FilterModules()`.

## Templates

- Remove an output that no longer applies with `GeneratePolicy: engine.RemoveIf(<predicate>, engine.PolicyDefault)`.
- A file the user also edits (*e.g.* `README.md`) uses `engine.GlobsWithPart` and `Sections` with `BEGIN_KICKR_<NAME>` and `END_KICKR_<NAME>` markers.
- Filter a module list once into a local var, and reuse it for both the guard and the loop.
- Never duplicate the filter predicate.
- Configure or disable a to-be-continuous component job through the component `inputs:`, never through `rules:` or `variables:` overrides.
- Omit a disabled hand-written job (no component) entirely, never write `when: never`.

### Standalone commands

- A standalone command includes only the templates its part owns, since a `RemoveIf` guard on a nil config deletes another platform's files.
- A standalone command behaves like the full generation for the parts it includes.

## Tests

- Generation tests in `pkg/generate/generate_test.go` compare against golden fixtures.
- The `test(ctx, t, repo, parsers...)` helper diffs the output against `testdata/<TestName>/...`.
- Build a case config with `merge(t, base, tc.Config)`.
- Reuse the existing `golang`, `node` and `hugo` parser closures.
- Stop and ask before refactoring or duplicating one, never implement a variant and explain it after.

## Keep in sync

- A new parser: `Parsers` in `pkg/generate/generate.go`, where order matters.
- A new generator or template set:
  - `Generators` in `pkg/generate/generate.go`
  - the standalone command owning that part, if any
- A new standalone command: `Execute` in `internal/cobra/root.go` and its `internal/cobra/<name>_test.go`.
- A template change: the fixtures under `testdata/`, which the user regenerates.

## Checks

The user runs every check.
Name the ones a change needs.

- Unit tests with `just test`.
- Lint with `just lint`.
- Golden fixtures with `just testdata`, after a template or generator change.
- pre-commit (gitleaks, whitespace) runs on every commit.

## Boundaries

- Never run `go` or `just` commands (build, lint, test, testdata).
- The only exception is the schema `go generate` in [Schema](#schema).
- Make the edits and stop.
- Ask before picking a side when a config combination is ambiguous or contradictory (*e.g.* website hosting vs Docker), even if a previous answer seemed to settle it.
