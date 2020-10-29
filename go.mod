module github.com/mildwonkey/terraform-provider-random

go 1.15

require (
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/hashicorp/go-plugin v1.3.0
	github.com/hashicorp/terraform-plugin-go v0.0.0-20201027121849-e227023a4d99
)

replace github.com/hashicorp/terraform-plugin-go => github.com/mildwonkey/terraform-plugin-go v0.0.0-20201029133533-aaa214ff26dc
