package app

import (
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

func NewRouter() (*mux.Router, error) {
	r := mux.NewRouter()
	r.HandleFunc("/version", versionHandler).Methods(http.MethodGet)
	return r, nil
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Ctx(ctx).Debug().Msg("versionHandler")
	resp := Response{
		Item: Version,
	}
	resp.Respond(w)
}
