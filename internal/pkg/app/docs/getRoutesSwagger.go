package docs

// swagger:route GET /v1/routes routes getRoutesId
// Gets list of all routes currently present in the routing table.
// responses:
//   200: routesResponse
//   500: routeErrorResponse

import "github.com/xmidt-org/ears/pkg/route"

// Items response containing a list of routes.
// swagger:response routesResponse
type routesResponseWrapper struct {
	// in:body
	Body routesResponse
}

type routesResponse struct {
	Status responseStatus `json:"status"`
	Items  []route.Config `json:"items"`
}
