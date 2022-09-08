module github.com/xmidt-org/ears

go 1.16

require (
	github.com/Shopify/sarama v1.36.0
	github.com/aws/aws-sdk-go v1.44.93
	github.com/boriwo/deepcopy v0.0.0-20220804211148-d5122121a902
	github.com/dop251/goja v0.0.0-20210912140721-ac5354e9a820
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-redis/redis/v8 v8.11.5
	github.com/goccy/go-yaml v1.9.5
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/go-cmp v0.5.8
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/onsi/gomega v1.20.2
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.25.0
	github.com/sebdah/goldie/v2 v2.5.3
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.13.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xorcare/pointer v1.1.0
	go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama v0.23.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.34.0
	go.opentelemetry.io/contrib/propagators v0.22.0
	go.opentelemetry.io/otel v1.9.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.23.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.23.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC3
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.23.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.9.0
	go.opentelemetry.io/otel/metric v0.28.0
	go.opentelemetry.io/otel/sdk v1.9.0
	go.opentelemetry.io/otel/sdk/export/metric v0.28.0
	go.opentelemetry.io/otel/sdk/metric v0.28.0
	go.opentelemetry.io/otel/trace v1.9.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/fx v1.18.1
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/net v0.0.0-20220809184613-07c6da5e1ced
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	gopkg.in/yaml.v2 v2.4.0
)
