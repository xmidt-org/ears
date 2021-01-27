package docs

// swagger:route GET /version version versionId
// version gets current version of API.
// responses:
//   200: versionResponse

// This text will appear as description of your response body.
// swagger:response versionResponse
type versionResponseWrapper struct {
	// in:body
	Body versionResponse
}

type responseStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type versionResponse struct {
	Status responseStatus `json:"status"`
	Item   string         `json:"item"`
}
