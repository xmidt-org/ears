# ears.yaml

```
# ears.yaml example config file 

ears:

  # environment and log level

  env: local
  
  logLevel: info

  # ears rest api port

  api:
    port: 3000
    
  # route and tenant storage  

  storage:
    route:
      type: inmemory
      #type: dynamodb
      region: us-west-2
      tableName: ears.routes.demo
    tenant:
      type: inmemory
      #type: dynamodb
      region: us-west-2
      tableName: ears.tenants.demo

  # routing table synchronization

  synchronization:
    type: inmemory
    #type: redis
    endpoint: localhost:6379
    active: no

  # optional rate limiter

  ratelimiter:
    type: inmemory
    #type: redis
    endpoint: localhost:6379
    active: yes

  # use otel collector for metrics and traces

  opentelemetry:
    otel-collector:
      active: no
      protocol: "grpc"
      endpoint: "localhost:55680"
    stdout:
      active: no

  secrets:

    # globally availabe secrets
    # access with secret://kafka.caCert for example

    all:
      all:
          kafka:
            caCert: |-
              -----BEGIN CERTIFICATE-----
              ...  
              -----END CERTIFICATE-----
            accessCert: |-
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
            accessKey: |-
              -----BEGIN PRIVATE KEY-----
              ...
              -----END PRIVATE KEY-----

            accessCertificate: |-
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----

    # tenant specific secrets
    # access with secret://kafka.caCert for example

    {orgid}:
      {appid}:
          kafka:
            caCert: |-
              -----BEGIN CERTIFICATE-----
              ...  
              -----END CERTIFICATE-----
            accessCert: |-
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
            accessKey: |-
              -----BEGIN PRIVATE KEY-----
              ...
              -----END PRIVATE KEY-----
            accessCertificate: |-
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
```