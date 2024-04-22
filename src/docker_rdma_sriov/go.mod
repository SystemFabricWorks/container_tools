module docker_rdma_sriov

go 1.22.2

replace google.golang.org/genproto/googleapis/api/httpbody => google.golang.org/genproto/googleapis/api/httpbody v0.0.0-20240227224415-6ceb2ff114de

replace google.golang.org/genproto/googleapis/rpc => google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de

replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de

require (
	github.com/Mellanox/rdmamap v1.1.1-0.20230906090736-d8230af2dc58
	github.com/docker/docker v26.0.2+incompatible
	github.com/k8snetworkplumbingwg/sriovnet v1.2.0
	github.com/spf13/cobra v1.8.0
	github.com/vishvananda/netlink v1.2.1-beta.2.0.20240411215012-578e95cc3190
	github.com/vishvananda/netns v0.0.5-0.20240412164733-9469873f4601
)

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/afero v1.9.4 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0 // indirect
	go.opentelemetry.io/otel v1.25.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.25.0 // indirect
	go.opentelemetry.io/otel/metric v1.25.0 // indirect
	go.opentelemetry.io/otel/sdk v1.25.0 // indirect
	go.opentelemetry.io/otel/trace v1.25.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
