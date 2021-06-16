package secret

import "github.com/xmidt-org/ears/pkg/tenant"

type Vault interface {
	//Secret returns a secret given a tenant Id and a key. Return an empty string if a secret is not found
	Secret(tid tenant.Id, key string) string
}
