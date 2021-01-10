init:
	go get -v ./...
	go get github.com/cespare/reflex

dev:
	reflex -d none -t 1s -s -R vendor. -r '\.(go|yml)$$' -- sh -c 'go run ./cmd'

protoc:
	protoc -I=internal/pkg/messages/ --go_out=internal/pkg/messages/ internal/pkg/messages/*.proto
	pbjs -t static-module -w commonjs -o web/events.js internal/pkg/messages/*.proto
