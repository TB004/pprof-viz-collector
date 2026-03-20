module github.com/TB004/pprof-viz-agent

go 1.26.1

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)

require (
	connectrpc.com/connect v1.19.1
	github.com/klauspost/compress v1.18.4
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.52.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
