.PHONY: dev
dev: build
	@mv ./kickr ~/.local/bin/kickr.dev

.PHONY: ua
ua:
	@./scripts/sh/update.sh

.PHONY: testdata
testdata:
	@TESTDATA=1 go test ./... -run ^TestGenerate_ -count 1 -timeout=15s
