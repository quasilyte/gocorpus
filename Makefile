.PHONY: all wasm ts run-server

run-server:
	go run _script/server.go

ts:
	tsc --target es6 app.ts

wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm

all: ts wasm
