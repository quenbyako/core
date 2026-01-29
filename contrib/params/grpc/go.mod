module github.com/quenbyako/core/contrib/params/grpc

go 1.25.1

replace github.com/quenbyako/core => ../../..

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.9-20250912141014-52f32327d4b0.1
	buf.build/go/protovalidate v1.0.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.2
	github.com/quenbyako/core v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250825161204-c5933d9347a5 // indirect
)
