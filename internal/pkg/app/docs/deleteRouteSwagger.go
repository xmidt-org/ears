package docs

// swagger:route DELETE /v1/routes/{routeId} routes deleteRoute
// Removes an existing route from the routing table if a route with the given ID exists.
// responses:
//   200: RouteDeleteResponse
//   500: RouteErrorResponse

// Item response containing the ID of the deleted route.
// swagger:response routeDeleteResponse
type routeDeleteResponseWrapper struct {
	// in: body
	Body RouteDeleteResponse
}

type RouteDeleteResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
