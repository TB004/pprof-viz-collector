# Prints all available receipes
help:
    - just --list

# Generate ConnectRPC code from protobuf and the runtime metric table
[group("miscellaneous")]
gen:
    buf generate
    go generate ./internal/...

# Fix field alignments in structs to improve performance
[group("miscellaneous")]
fix_alignments:
    fieldalignment -fix ./...

# Check source code for vulnerabilities
[group("maintain")]
vuln:
    staticcheck ./...
    govulncheck ./...
    gosec -enable-audit -exclude-generated ./...

# Run tests with coverage
[group("maintain")]
cover:
    go test -race -cover -coverprofile cover.out  -outputdir ./.test ./...
    go tool cover -html=./.test/cover.out  -o ./.test/coverage.html

# Run tests
[group("maintain")]
test:
    go test -race ./...

# Run benchmarks with allocation reporting
[group("maintain")]
bench:
    go test ./internal/... -bench . -benchmem -run '^$'

# Runs different tool to lint the project
[group("maintain")]
lint:
    go vet ./...
    golangci-lint run ./...
    modernize ./...
    fieldalignment ./...

# Install necessary go tools
[group("setup")]
install:
    go install github.com/cweill/gotests/gotests@latest
    go install github.com/go-delve/delve/cmd/dlv@latest
    go install github.com/josharian/impl@latest
    go install golang.org/x/tools/gopls@latest
    go install honnef.co/go/tools/cmd/staticcheck@latest
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    go install golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest
    go install github.com/bufbuild/buf/cmd/buf@latest
    go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
