build-window:
	GOOS=windows GOARCH=amd64 go build -o crawlify.exe ./cmd/main.go

PHONY: build-window
