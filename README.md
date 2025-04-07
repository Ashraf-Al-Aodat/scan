## Goal:
the goal is to create a LLM tool that can scan repos to see if there is any security vulnerability in the code.
## Usage
```console
go mod tidy # run once
go run cmd/analyzer/main.go  -p {path/to/local/repo/}
```
