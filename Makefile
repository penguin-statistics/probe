init:
	go get -v ./...
	go get github.com/cespare/reflex

dev:
	export PENGUINPROBE_DSN="host=localhost user=root password=root dbname=penguinprobe port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	export PENGUINPROBE_DEBUG="true"
	reflex -d none -t 1s -s -R vendor. -r '\.(go|yml)$$' -- sh -c 'go run ./cmd'

protoc:
	protoc -I=internal/pkg/messages/ --go_out=internal/pkg/messages/ internal/pkg/messages/*.proto
	pbjs -t static-module -w commonjs -o web/events.js internal/pkg/messages/*.proto
