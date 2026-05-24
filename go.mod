module github.com/TB004/pprof-viz-agent

go 1.26.2

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)

require (
	connectrpc.com/connect v1.20.0
	connectrpc.com/cors v0.1.0
	github.com/klauspost/compress v1.18.6
	github.com/quic-go/quic-go v0.59.0
	github.com/rs/cors v1.11.1
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.53.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
