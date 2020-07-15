module github.com/openebs/node-disk-manager

go 1.14

require (
	github.com/diskfs/go-diskfs v1.1.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/golang/protobuf v1.4.2
	github.com/mitchellh/go-ps v1.0.0
	github.com/onsi/ginkgo v1.12.3
	github.com/onsi/gomega v1.10.1
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/prometheus/client_golang v1.5.1
	github.com/spf13/cobra v0.0.7
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.23.0
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.2
)

replace k8s.io/client-go => k8s.io/client-go v0.17.4
