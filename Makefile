init:
	go install github.com/mitranim/gow@latest

dev:
	gow -c -e go,yml,mod run ./cmd

protoc:
	protoc -I=internal/pkg/messages/ --go_out=internal/pkg/messages/ internal/pkg/messages/*.proto
	pbjs -t static-module -w commonjs -o web/events.js internal/pkg/messages/*.proto
