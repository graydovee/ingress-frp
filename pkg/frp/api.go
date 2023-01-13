package frp

import "net/http"

//subRouter.HandleFunc("/api/reload", svr.apiReload).Methods("GET")
//subRouter.HandleFunc("/api/status", svr.apiStatus).Methods("GET")
//subRouter.HandleFunc("/api/config", svr.apiGetConfig).Methods("GET")
//subRouter.HandleFunc("/api/config", svr.apiPutConfig).Methods("PUT")

type API interface {
	Method() string
	URI() string
}

type api struct {
	uri    string
	method string
}

func (a *api) Method() string {
	return a.method
}

func (a *api) URI() string {
	return a.uri
}

var (
	ApiGetConfig API = &api{uri: "/api/config", method: http.MethodGet}
	ApiPutConfig API = &api{uri: "/api/config", method: http.MethodPut}
	ApiReload    API = &api{uri: "/api/reload", method: http.MethodGet}
)
