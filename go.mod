module github.com/ebauman/moo

go 1.15

require (
	github.com/golang/protobuf v1.4.1
	github.com/google/martian v2.1.0+incompatible
	github.com/hashicorp/go-uuid v1.0.1
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/rancher/norman v0.0.0-20200517050325-f53cae161640
	github.com/rancher/types v0.0.0-20200326224903-b4612bd96d9b
	github.com/sirupsen/logrus v1.6.0
	github.com/terraform-providers/terraform-provider-rancher2 v1.8.3
	github.com/urfave/cli/v2 v2.2.0
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.24.0
	k8s.io/apimachinery v0.18.0
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.0
