package docs

// swagger:route POST /v1/routes routes postRoute
// Adds a new route to the routing table or updates an existing route. Route ID can be given in the body. If it is omitted a hash will be calculates and used instead.
// responses:
//   200: RouteResponse
//   500: RouteErrorResponse
