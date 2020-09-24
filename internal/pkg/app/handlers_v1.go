package app

import (
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

type APIManager struct {
	muxRouter       *mux.Router
	routingTableMgr RoutingTableManager
}

func NewAPIManager(routingMgr RoutingTableManager) (*APIManager, error) {
	api := &APIManager{
		muxRouter:       mux.NewRouter(),
		routingTableMgr: routingMgr,
	}
	api.muxRouter.HandleFunc("/version", api.versionHandler).Methods(http.MethodGet)
	api.muxRouter.HandleFunc("/v1/routes/{route}", api.addRouteHandler).Methods(http.MethodPut)
	return api, nil
}

func (a *APIManager) versionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Ctx(ctx).Debug().Msg("versionHandler")
	resp := Response{
		Item: Version,
	}
	resp.Respond(w)
}

func (a *APIManager) addRouteHandler(w http.ResponseWriter, r *http.Request) {
	//TODO this handler is incomplete. It is here for demo purpose
	ctx := r.Context()
	err := a.routingTableMgr.AddRoute(ctx, nil)
	if err != nil {
		log.Ctx(ctx).Error().Str("op", "addRouteHandler").Msg(err.Error())
	}
	resp := Response{
		Status: &Status{
			Code:    http.StatusNotImplemented,
			Message: "This method is not yet implemented",
		},
	}
	resp.Respond(w)
}
