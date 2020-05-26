module github.com/ebauman/moo

go 1.13

require (
	github.com/google/martian v2.1.0+incompatible
	github.com/rancher/norman v0.0.0-20200517050325-f53cae161640
	github.com/rancher/types v0.0.0-20200326224903-b4612bd96d9b
	github.com/sirupsen/logrus v1.6.0
	github.com/terraform-providers/terraform-provider-rancher2 v1.8.3
	github.com/urfave/cli/v2 v2.2.0
	k8s.io/api v0.18.0
	k8s.io/apimachinery v0.18.0
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.0
