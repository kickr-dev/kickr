# kickr

## Schema and generated Go types

The schema definition `.schemas/kickr.v1.schema.yml` is source of truth.
Never edit `.schemas/kickr.v1.schema.json` or `pkg/kickr/v1/kickr.go`/`constants.go` directly:
1. Edit the yml
2. Run `go run ./schemas/gen/main.go`
3. Use the `schema-to-go` skill to update `kickr.go` (structs/GoDoc) and `constants.go` (enums).

## Types

### Methods

- Path: `pkg/generate/types`
- No language-specific methods on `Repository`/`Module` (e.g. `HasHugo()`). Language logic goes in templates or a dedicated parser.
- No package-level lookup-list `var`s; declare locally in the function that needs them.
- Reuse existing helpers (`Module.HasDocker()`, `Module.IsWebsite()`, `Repository.IsHosting()`, `Repository.ModulesWith()`) instead of hand-rolled comparisons.

## Templates

- Under `pkg/generate/templates/_templates`
- Filter a module list once into a local var, reuse for both guard and loop. Never duplicate the filter predicate.
- to-be-continuous component jobs: disable/configure via component `inputs:`, not `rules:`/`variables:` overrides.
- Hand-written jobs (no component): omit entirely if disabled, no `when: never`.

## Tests

### Generation tests

- Path: `pkg/generate/generate_test.go`
- Golden-fixture: `test(ctx, t, repo, parsers...)` diffs against `testdata/<TestName>/...`.
- Regenerate fixtures: `make testdata`.
- Reuse existing `golang`/`node`/`hugo` parser closures instead of duplicating.

---

@README.md
