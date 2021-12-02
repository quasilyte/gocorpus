.PHONY: all wasm ts

ts:
	tsc --target es6 app.ts

wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm

all: ts wasm
