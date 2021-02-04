package docs

// swagger:route GET /version version version
// version gets current version of API.
// responses:
//   200: VersionResponse

// This text will appear as description of your response body.
// swagger:response versionResponse
type versionResponseWrapper struct {
	// in: body
	Body VersionResponse
}

type responseStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type VersionResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
