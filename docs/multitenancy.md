# EARS Multitenancy

In EARS, org IDs and application IDs are used to support multi-tenancy. Conceptually, one can think that an org ID contains multiple application IDs, and an application ID contains multiple routes such that:

* An org ID must be unique globally.
* An application ID needs to be unique within an org ID.
* A route ID needs to be unique within an org ID-application ID pair.

The org-application ID pair forms the concept of a tenant that provides logical separations for routes such that routes in one tenant has no visibility of routes in another tenant. There should be minimal performance implication between routes in different tenants. For example, a bad acting route in one tenant may affect performance of routes in the same tenant, but not routes in another tenant.

The org-application ID pair also allows for quota and rate-limit at the tenant level (more on this later)

## EARS route REST APIs

When configuring routes through EARS REST API, a user must specified the org ID and the application ID in the request path. 

For example, to create a route using PUT:
```
PUT /ears/v1/org/{orgId}/applications/{appId}/routes/{routeId}
```

Org ID and application ID must exist before we can create routes on them. 

CRUD APIs for org ID:
```
PUT    /ears/v1/org/{orgId}  //create an OrgId
GET    /ears/v1/org/{orgId}  //get info about an OrgId
DELETE /ears/v1/org/{orgId}  //delete an OrgId
```

CRUD APIs for application ID:
```
PUT    /ears/v1/org/{orgId}/applications/{appId}  //greate an application ID
GET    /ears/v1/org/{orgId}/applications/{appId}  //get info about an application ID
DELETE /ears/v1/org/{orgId}/applications/{appId}  //delete an application ID
```

The `route.Config` struct has a `TenantId` field. These are for internal use only.

## EARS route storer interface

Current implementation:
```go
type RouteStorer interface {
	//If route is not found, the function should return a item not found error
	GetRoute(context.Context, string) (Config, error)
	GetAllRoutes(context.Context) ([]Config, error)

	//SetRoute will add the route if it new or update the route if
	//it is an existing one. It will also update the create time and
	//modified time of the route where appropriate.
	SetRoute(context.Context, Config) error

	SetRoutes(context.Context, []Config) error

	DeleteRoute(context.Context, string) error

	DeleteRoutes(context.Context, []string) error
}

type Config struct {
    Id           string         `json:"id,omitempty"`    // route ID
    OrgId        string         `json:"orgId,omitempty"` // org ID for quota and rate limiting
    AppId        string         `json:"appId,omitempty"` // app ID for quota and rate limiting
    ...
}
```

Proposed Change:
```go
type TenantId interface {
    AppId() string
    OrgId() string
    Hash() string
}

type Config struct {
    Id  string    `json:"id,omitempty"`    // route ID
    tId TenantId  //internal fields, cannot be marshalled with the default json marshaller       
    ...
}

type RouteStorer interface {
	//If route is not found, the function should return a item not found error
	GetRoute(context.Context, tid TenantId, id string) (Config, error)
	GetAllRoutes(context.Context, tid TenantId) ([]Config, error)

	//SetRoute will add the route if it new or update the route if
	//it is an existing one. It will also update the create time and
	//modified time of the route where appropriate.
	SetRoute(context.Context, Config) error

	SetRoutes(context.Context, []Config) error

	DeleteRoute(context.Context, tid TenantId, id string) error

	DeleteRoutes(context.Context, tid TenantId, ids []string) error
}


```
For the getters, does the storer needs to populate the tenant id in the returning route.Confg?

## Related code updates
* tenant ID at plugin level 
* route uniqueness needs to be based on a combo of (tenant + route id) for inmemory route hashtable

* Make route ID its own type 