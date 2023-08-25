package secret

import "context"

const ProtocolSecret = "secret://"
const ProtocolCredential = "credential://"

type Vault interface {
	//Secret returns a secret given a key. Return an empty string if a secret is not found
	Secret(ctx context.Context, key string) string
	GetConfig(ctx context.Context, key string) string
}
