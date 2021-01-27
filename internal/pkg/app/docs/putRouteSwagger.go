package docs

// swagger:route PUT /v1/routes/{routeId} routes putRoute
// Adds a new route to the routing table or updates an existing route. Route ID can be given in the body. If it is omitted a hash will be calculates and used instead.
// responses:
//   200: routeResponse
//   500: routeErrorResponse

import "github.com/xmidt-org/ears/pkg/route"

// Item response containing a complete route configuration including sender, receiver and optional filter chain.
// swagger:response routeResponse
type routeResponseWrapper struct {
	// in: body
	Body routeResponse
}

// Item response containing a route error.
// swagger:response routeErrorResponse
type routeErrorResponseWrapper struct {
	// in: body
	Body routeErrorResponse
}

// swagger:parameters putRoute postRoute
type routeParamWrapper struct {
	// Route configuration including sender, receiver and optional filter chain.
	// in: body
	// required: true
	Body route.Config
}

// swagger:parameters putRoute getRoute deleteRoute
type routeIdParamWrapper struct {
	// Route ID
	// in: path
	// required: true
	RouteId string `json:"routeId"`
}

type routeResponse struct {
	Status responseStatus `json:"status"`
	Item   route.Config   `json:"item"`
}

type routeErrorResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
