module github.com/opencord/bbsim-sadis-server

go 1.16

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
)

require (
	github.com/gorilla/mux v1.8.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/opencord/voltha-lib-go/v7 v7.0.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
)
