package main

import (
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/ghpr"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/runtimelogs"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

var logger, _ = zap.NewProduction()
var sugar = logger.Sugar()


// Routes sets up the router and mounts the routes.
func Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json
		middleware.Logger,          // Log API request calls
		middleware.DefaultCompress, // Compress results, mostly gzipping assets and json
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
	)

	router.Route("/v1", func(r chi.Router) {
		r.Mount("/api/prcoverage", RoutesPR())
	})

	return router
}

type PRPayload struct {
	prID int `json:"prID"`
	repoURL string `json:"repoURL"`
}

func RoutesPR() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", RunPRCoverage)
	return router
}

func RunPRCoverage(w http.ResponseWriter, r *http.Request) {
	// Read body
	msg := PRPayload{
		prID: 482,
		repoURL: "openshift/machine-config-operator/",
	}
	_, _, sha := ghpr.GetPRPayload(msg.repoURL, msg.prID, "/tmp")
	//err := ioutil.WriteFile("/tmp/e2e_log.log", runtimeLogs, 0777)
	//if err != nil {
	//	sugar.Errorf("%v\n", err)
	//}
	runtimeCodePaths := runtimelogs.CreateRuntimePaths("/tmp/e2e_log.log")
	gremlin.AddRuntimePathsToGraph("machine-config-operator", sha, runtimeCodePaths)
	render.JSON(w, r, msg) // Return the same thing for now.
}

func main() {
	router := Routes()

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		sugar.Infof("%s %s\n", method, route) // Walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		sugar.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	sugar.Info(http.ListenAndServe(":8080", router)) // Note, the port is usually gotten from the environment.
}
