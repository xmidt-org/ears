package docs

// swagger:route GET /v1/routes routes getRoutes
// Gets list of all routes currently present in the routing table.
// responses:
//   200: RoutesResponse
//   500: RouteErrorResponse

import "github.com/xmidt-org/ears/pkg/route"

// Items response containing a list of routes.
// swagger:response routesResponse
type routesResponseWrapper struct {
	// in: body
	Body RoutesResponse
}

type RoutesResponse struct {
	Status responseStatus `json:"status"`
	Items  []route.Config `json:"items"`
}
