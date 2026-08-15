# build kickr and move it into ~/local/bin/kickr.dev.
dev: build
  @mv ./kickr ~/.local/bin/kickr.dev

# update all local repositories with a kickr configuration using either kickr.dev or kickr binary.
ua:
  @./scripts/sh/update.sh

# update all testdata files with current Go updated code (no binary involved).
testdata:
  @TESTDATA=1 go test ./... -run ^TestGenerate_ -count 1 -timeout=15s
