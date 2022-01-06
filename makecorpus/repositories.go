package main

type repository struct {
	name     string
	tags     []string
	git      string
	srcRoots []string
}

var repositoryList = []*repository{
	{
		name:     "goroot",
		tags:     []string{"compiler", "parser", "lib", "go-tools"},
		git:      "https://github.com/golang/go.git",
		srcRoots: []string{"src"},
	},
	{
		name:     "gopherjs",
		tags:     []string{"compiler", "parser"},
		git:      "https://github.com/gopherjs/gopherjs.git",
		srcRoots: []string{"compiler"},
	},
	{
		name:     "x-tools",
		tags:     []string{"lib", "go-tools"},
		git:      "https://github.com/golang/tools.git",
		srcRoots: []string{"."},
	},
	{
		name:     "protobuf",
		tags:     []string{"compiler", "encoder", "decoder"},
		git:      "https://github.com/golang/protobuf.git",
		srcRoots: []string{"."},
	},
	{
		name:     "sirupsen-logrus",
		tags:     []string{"lib", "logging"},
		git:      "https://github.com/sirupsen/logrus.git",
		srcRoots: []string{"."},
	},
	{
		name:     "uber-zap",
		tags:     []string{"lib", "logging"},
		git:      "https://github.com/uber-go/zap.git",
		srcRoots: []string{"."},
	},
	{
		name:     "x-net",
		tags:     []string{"lib", "net"},
		git:      "https://github.com/golang/net.git",
		srcRoots: []string{"."},
	},
	{
		name:     "x-text",
		tags:     []string{"lib"},
		git:      "https://github.com/golang/text.git",
		srcRoots: []string{"."},
	},
	{
		name:     "x-crypto",
		tags:     []string{"lib", "crypto", "math"},
		git:      "https://github.com/golang/crypto.git",
		srcRoots: []string{"."},
	},
	{
		name:     "spf13-cobra",
		tags:     []string{"lib", "cli"},
		git:      "https://github.com/spf13/cobra.git",
		srcRoots: []string{"."},
	},
	{
		name:     "go-kit",
		tags:     []string{"framework", "net"},
		git:      "https://github.com/go-kit/kit.git",
		srcRoots: []string{"."},
	},
	{
		name:     "valyala-fasthttp",
		tags:     []string{"lib", "net"},
		git:      "https://github.com/valyala/fasthttp.git",
		srcRoots: []string{"."},
	},
	{
		name:     "gorilla-websocket",
		tags:     []string{"lib", "net"},
		git:      "https://github.com/gorilla/websocket.git",
		srcRoots: []string{"."},
	},
	{
		name:     "gorilla-mux",
		tags:     []string{"lib", "net"},
		git:      "https://github.com/gorilla/mux.git",
		srcRoots: []string{"."},
	},
	{
		name:     "stretchr-testify",
		tags:     []string{"lib", "testing"},
		git:      "https://github.com/stretchr/testify.git",
		srcRoots: []string{"."},
	},
	{
		name:     "julienschmidt-httprouter",
		tags:     []string{"lib", "net"},
		git:      "https://github.com/julienschmidt/httprouter.git",
		srcRoots: []string{"."},
	},
	{
		name:     "jmoiron-sqlx",
		tags:     []string{"lib", "sql"},
		git:      "https://github.com/jmoiron/sqlx.git",
		srcRoots: []string{"."},
	},
	{
		name:     "json-iterator",
		tags:     []string{"lib", "decoder"},
		git:      "https://github.com/json-iterator/go.git",
		srcRoots: []string{"."},
	},
	{
		name:     "gonum",
		tags:     []string{"lib", "math"},
		git:      "https://github.com/gonum/gonum.git",
		srcRoots: []string{"."},
	},
	{
		name:     "goccy-yaml",
		tags:     []string{"lib", "parser"},
		git:      "https://github.com/goccy/go-yaml.git",
		srcRoots: []string{"."},
	},
	{
		name:     "gorm",
		tags:     []string{"lib", "orm", "sql"},
		git:      "https://github.com/go-gorm/gorm.git",
		srcRoots: []string{"."},
	},
	{
		name:     "xorm",
		tags:     []string{"lib", "orm", "sql"},
		git:      "https://gitea.com/xorm/xorm.git",
		srcRoots: []string{"."},
	},
	{
		name:     "kubernetes",
		tags:     []string{"tool"},
		git:      "https://github.com/kubernetes/kubernetes.git",
		srcRoots: []string{"."},
	},
	{
		name:     "helm",
		tags:     []string{"tool"},
		git:      "https://github.com/helm/helm.git",
		srcRoots: []string{"."},
	},
	{
		name:     "moby",
		tags:     []string{"tool"},
		git:      "https://github.com/moby/moby.git",
		srcRoots: []string{"."},
	},
	{
		name:     "hugo",
		tags:     []string{"tool"},
		git:      "https://github.com/gohugoio/hugo.git",
		srcRoots: []string{"."},
	},
	{
		name:     "traefik",
		tags:     []string{"tool"},
		git:      "https://github.com/traefik/traefik.git",
		srcRoots: []string{"."},
	},
	{
		name:     "caddy",
		tags:     []string{"tool"},
		git:      "https://github.com/caddyserver/caddy.git",
		srcRoots: []string{"."},
	},
	{
		name:     "minio",
		tags:     []string{"tool"},
		git:      "https://github.com/minio/minio.git",
		srcRoots: []string{"."},
	},
	{
		name:     "hashicorp-terraform",
		tags:     []string{"tool"},
		git:      "https://github.com/hashicorp/terraform.git",
		srcRoots: []string{"."},
	},
	{
		name:     "hashicorp-nomad",
		tags:     []string{"tool"},
		git:      "https://github.com/hashicorp/nomad.git",
		srcRoots: []string{"."},
	},
	{
		name:     "gitea",
		tags:     []string{"tool"},
		git:      "https://github.com/go-gitea/gitea.git",
		srcRoots: []string{"."},
	},
	{
		name:     "golangci-lint",
		tags:     []string{"tool"},
		git:      "https://github.com/golangci/golangci-lint.git",
		srcRoots: []string{"."},
	},
	{
		name:     "etcd",
		tags:     []string{"db"},
		git:      "https://github.com/etcd-io/etcd.git",
		srcRoots: []string{"."},
	},
	{
		name:     "cockroach",
		tags:     []string{"db"},
		git:      "https://github.com/cockroachdb/cockroach.git",
		srcRoots: []string{"."},
	},
	{
		name:     "dgraph-badger",
		tags:     []string{"db"},
		git:      "https://github.com/dgraph-io/badger.git",
		srcRoots: []string{"."},
	},

	{
		name:     "victoriametrics",
		tags:     []string{"db", "metrics"},
		git:      "https://github.com/VictoriaMetrics/VictoriaMetrics.git",
		srcRoots: []string{"app", "lib"},
	},
	{
		name:     "jackc-pgx",
		tags:     []string{"lib", "sql"},
		git:      "https://github.com/jackc/pgx.git",
		srcRoots: []string{"."},
	},
	{
		name:     "prometheus",
		tags:     []string{"db", "metrics"},
		git:      "https://github.com/prometheus/prometheus.git",
		srcRoots: []string{"."},
	},
	{
		name:     "talos",
		tags:     []string{"os"},
		git:      "https://github.com/talos-systems/talos.git",
		srcRoots: []string{"."},
	},
	{
		name:     "drone",
		tags:     []string{"ci"},
		git:      "https://github.com/harness/drone.git",
		srcRoots: []string{"."},
	},
	{
		name:     "grafana",
		tags:     []string{"metrics"},
		git:      "https://github.com/grafana/grafana.git",
		srcRoots: []string{"."},
	},

	{
		name:     "influxdb",
		tags:     []string{"db"},
		git:      "https://github.com/influxdata/influxdb.git",
		srcRoots: []string{"."},
	},
	{
		name:     "vitess",
		tags:     []string{"db"},
		git:      "https://github.com/vitessio/vitess.git",
		srcRoots: []string{"."},
	},
	{
		name:     "cilium",
		tags:     []string{"ebpf"},
		git:      "https://github.com/cilium/cilium.git",
		srcRoots: []string{"."},
	},
	{
		name:     "rook",
		tags:     []string{"kubernetes"},
		git:      "https://github.com/rook/rook.git",
		srcRoots: []string{"."},
	},
	{
		name:     "tyk",
		tags:     []string{"grpc"},
		git:      "https://github.com/TykTechnologies/tyk.git",
		srcRoots: []string{"."},
	},
	{
		name:     "grpc-go",
		tags:     []string{"grpc", "net", "lib"},
		git:      "https://github.com/grpc/grpc-go.git",
		srcRoots: []string{"."},
	},
}
