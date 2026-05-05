.PHONY: generate build dev clean

generate:
	templ generate

build: generate
	go build ./...

dev:
	templ generate --watch & go run cmd/server/main.go

clean:
	find . -name "*_templ.go" -delete
	find . -name "*.templ.txt" -delete
