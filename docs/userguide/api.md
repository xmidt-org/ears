# EARS REST API

Find the full Open API spec [here](../../internal/pkg/app/swagger.yaml). Use go-swagger to 
view as HTML:

```
swagger serve -F=swagger swagger.yaml
```

## Tenant CRUD Operations

EARS supports multi-tenancy and is therefore suitable to be offered as a service.
This means, before you can create a route for a tenant, you must create that tenant
with a valid tenant configuration. A tenant is a two-dimensional structure consisting
of an org ID and an application ID. Currently, the only required configuration for
a tenant is its event throughput quota in events per second. EARS will then enforce
rate limiting accordingly, provided the rate limit feature is turned on in ears.yaml.

Example tenant configuration:

```
{
  "quota": {
    "eventsPerSec": 100
  }
}
```

### Get Tenant

```
GET /ears/v1/orgs/{orgId}/applications/{appId}/config
```

### Add / Update Tenant

```
PUT /ears/v1/orgs/{orgId}/applications/{appId}/config {tenantBody}
```

### Delete Tenant

```
DELETE /ears/v1/orgs/{orgId}/applications/{appId}/config
```

## Route CRUD Operations

Example route configuration:

```
{
  "id": "ks1",
  "userId": "boris",
  "name": "kafkaSqsRoute",
  "receiver": {
    "plugin": "kafka",
    "name": "kafkaSqsReceiver",
    "config": {
      "brokers": "localhost:9092",
      "topic": "ears-kafka-test",
      "groupId": "mygroup"
    }
  },
  "sender": {
    "plugin": "sqs",
    "name": "kafkaSqsSender",
    "config": {
      "queueUrl": "https://sqs.us-west-2.amazonaws.com/{accountId}/ears-sqs-test"
    }
  },
  "deliveryMode": ""
}
```

The above example directly connects a Kafka topic called _ears-kafka-test_ to an SQS queue called
_ears-sqs-test_. Notice that no filter chain is configured in this example so all events received from 
Kafka will be routed to SQS unmodified and unfiltered. Notice that the route has an ID and an optional 
name field. The ID must match the ID given in a PUT call or be blank. When no ID is given in a POST 
call a random route ID will be generated for you and returned with the API response. The _deliveryMode_
field is currently unused.

### Get Route

```
GET /ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}
```

### Update Route

```
PUT /ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId} {routeBody}
```

### Delete Route

```
DELETE /ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}
```

### Add Route

```
POST /ears/v1/orgs/{orgId}/applications/{appId}/routes {routeBody}
```

### Get All Route For Tenant

```
GET /ears/v1/orgs/{orgId}/applications/{appId}/routes
```

### Send Single Event To Route

```
POST /ears/v1/orgs/{orgId}/applications/{appId}/routes/{routeId}/event {eventBody}
```

## Admin APIs

### Get All Routes

Get all routes across all tenants.

```
GET /ears/v1/routes
```

### Get All Tenants

Get all tenant configurations.

```
GET /ears/v1/tenants
```

### Get All Senders

Get all sender plugins configurations across all routes and all tenants. A reference count is given in the
response to indicate by how many routes this plugin is shared.

```
GET /ears/v1/senders
```

### Get All Receivers

Get all receiver plugins configurations across all routes and all tenants. A reference count is given in the
response to indicate by how many routes this plugin is shared.

```
GET /ears/v1/receivers
```

### Get All Filters

Get all filter plugins configurations across all routes and all tenants. A reference count is given in the
response to indicate by how many routes this plugin is shared.

```
GET /ears/v1/filters
```
