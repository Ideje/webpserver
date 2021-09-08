package server

import (
	"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"webpserver/app"
	"webpserver/config"
	middleware "webpserver/middleware"
)

func Router() http.Handler {

	router := mux.NewRouter()

	router.Handle("/metrics", promhttp.Handler())
	router.PathPrefix("/").Handler(WebpServer())

	// the 404 default server is also going thru all the middlewares
	router.NotFoundHandler = router.NewRoute().HandlerFunc(http.NotFound).GetHandler()

	if app.LogLevel >= log15.LvlWarn {
		router.Use(middleware.LogMiddleware)
	}

	return router
}

func NewHTTPServer() *http.Server {
	handler := Router()

	server := &http.Server{
		Addr:    config.Cfg.ServerAddress,
		Handler: handler,
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return server
}
