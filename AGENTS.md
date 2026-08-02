# Kickr

## Schema and generated Go types

`.schemas/kickr.v1.schema.yml` is source of truth. Never edit `.schemas/kickr.v1.schema.json` or
`pkg/kickr/v1/kickr.go`/`constants.go` directly: edit the yml, run `go generate ./...`,
then mirror via `schema-to-go` skill into `kickr.go` (structs/GoDoc) and `constants.go` (enums).

## Repository/Module types (`pkg/generate/types`)

- No language-specific methods on `Repository`/`Module` (e.g. `HasHugo()`).
  Language logic goes in templates or a dedicated parser.
- No package-level lookup-list `var`s; declare locally in the function that needs them.
- Reuse existing helpers (`Module.HasDocker()`, `Module.IsWebsite()`, `Repository.IsHosting()`, `Repository.ModulesWith()`) instead of hand-rolled comparisons.

## Templates (`pkg/generate/templates/_templates`)

- Filter a module list once into a local var, reuse for both guard and loop.
  Never duplicate the filter predicate.
- to-be-continuous component jobs: disable/configure via component `inputs:`, not `rules:`/`variables:` overrides.
- Hand-written jobs (no component): omit entirely if disabled, no `when: never`.

## Tests (`pkg/generate/generate_test.go`)

- Golden-fixture: `test(ctx, t, repo, parsers...)` diffs against `testdata/<TestName>/...`.
- Regenerate fixtures: `make testdata`.
- Reuse existing `golang`/`node`/`hugo` parser closures instead of duplicating.

---

@README.md
