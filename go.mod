module watchmen

go 1.13

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.44.3
	cloud.google.com/go/datastore => github.com/googleapis/google-cloud-go/datastore v1.0.0
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/exp => github.com/golang/exp v0.0.0-20190731235908-ec7cb31e5a56
	golang.org/x/image => github.com/golang/image v0.0.0-20190802002840-cff245a6509b
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190409202823-959b441ac422
	golang.org/x/mobile => github.com/golang/mobile v0.0.0-20190814143026-e8b3e6111d02
	golang.org/x/mod => github.com/golang/mod v0.1.0
	golang.org/x/net => github.com/golang/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190813064441-fde4db37ae7a
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/time => github.com/golang/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools => github.com/golang/tools v0.0.0-20190816200558-6889da9d5479
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20190717185122-a985d3407aa7
	google.golang.org/api => github.com/googleapis/google-api-go-client v0.8.0
	google.golang.org/appengine => github.com/golang/appengine v1.6.1
	google.golang.org/genproto => github.com/googleapis/go-genproto v0.0.0-20190817000702-55e96fffbd48
	google.golang.org/grpc => github.com/grpc/grpc-go v1.23.0
	gopkg.in/inf.v0 => github.com/go-inf/inf v0.9.1
	k8s.io/api => github.com/kubernetes/api v0.0.0-20190814101207-0772a1bdf941
	k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.0.0-20190814100815-533d101be9a6
	k8s.io/client-go => github.com/kubernetes/client-go v0.0.0-20190805141520-2fe0317bcee0
	k8s.io/gengo => github.com/kubernetes/gengo v0.0.0-20190813173942-955ffa8fcfc9
	k8s.io/klog => github.com/kubernetes/klog v0.4.0
	k8s.io/kube-openapi => github.com/kubernetes/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/utils => github.com/kubernetes/utils v0.0.0-20190809000727-6c36bc71fc4a
	sigs.k8s.io/structured-merge-diff => github.com/kubernetes-sigs/structured-merge-diff v0.0.0-20190817042607-6149e4549fca
	sigs.k8s.io/yaml => github.com/kubernetes-sigs/yaml v1.1.0
	sum.golang.org/lookup/github.com/hashicorp/memberlist => github.com/hashicorp/memberlist v0.1.5
)

require (
	github.com/gin-gonic/gin v1.4.0
	github.com/go-kit/kit v0.9.0
	github.com/hashicorp/memberlist v0.1.5
	github.com/prometheus/client_golang v1.2.0
	github.com/sirupsen/logrus v1.4.2
	k8s.io/apimachinery v0.0.0-20190814100815-533d101be9a6
	k8s.io/client-go v0.0.0-00010101000000-000000000000
)
