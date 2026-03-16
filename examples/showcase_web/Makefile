GOROOT := $(shell go env GOROOT)
WASM_EXEC := $(GOROOT)/lib/wasm/wasm_exec.js

.PHONY: build serve clean

build: showcase.wasm wasm_exec.js

showcase.wasm: main.go ../../*.go
	GOOS=js GOARCH=wasm go build -o showcase.wasm .

wasm_exec.js:
	cp "$(WASM_EXEC)" .

serve: build
	@echo "Serving at http://localhost:8080"
	@python3 -m http.server 8080

clean:
	rm -f showcase.wasm wasm_exec.js
