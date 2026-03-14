export GOEXPERIMENT := "dwarf5,greenteagc,newinliner,jsonv2,loopvar,fieldtrack,randomizedheapbase64,runtimefreegc,sizespecializedmalloc,goroutineleakprofile"

# Prints all available receipes
help:
    - just --list

# Generate ConnectRPC code from protobuf
[group("miscellaneous")]
gen:
    - buf generate

# Check source code for vulnerabilities
[group("maintain")]
vuln:
    - staticcheck ./...
    - govulncheck ./...
    - gosec -enable-audit -exclude-generated -no-fail ./...

# Run tests with coverage
[group("maintain")]
cover:
    - go test -race -cover -coverprofile cover.out  -outputdir ./.test ./internal/...
    - go tool cover -html=./.test/cover.out  -o ./.test/coverage.html

# Run tests
[group("maintain")]
test:
    - go test -race ./...

# Runs the golangci-lint tool to lint the project
[group("maintain")]
lint:
    - go vet ./...
    - golangci-lint run ./...
    - modernize ./...

# Install necessary go tools
[group("setup")]
install:
    - go install github.com/cweill/gotests/gotests@latest
    - go install github.com/go-delve/delve/cmd/dlv@latest
    - go install github.com/josharian/impl@latest
    - go install golang.org/x/tools/gopls@latest
    - go install honnef.co/go/tools/cmd/staticcheck@latest
    - go install github.com/securego/gosec/v2/cmd/gosec@latest
    - go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    - go install golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest
    - go install golang.org/x/vuln/cmd/govulncheck@latest
    - go install github.com/bufbuild/buf/cmd/buf@latest
