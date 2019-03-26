package main

import (
	"io/ioutil"
	"net/http"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/runtimelogs"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/ghpr"

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
	prID    int    `json:"pr_id"`
	repoURL string `json:"repo_url"`
}

func RoutesPR() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", RunPRCoverage)
	return router
}

func RunPRCoverage(w http.ResponseWriter, r *http.Request) {
	// Read body
	msg := PRPayload{
		prID:    482,
		repoURL: "openshift/machine-config-operator/",
	}
	hunks, branchDetails, clonePath := ghpr.GetPRPayload(msg.repoURL, msg.prID, "/tmp")

	// Run the E2E tests on the cloned fork and write results to file.
	logFileE2E := runtimelogs.RunE2ETestsInGoPath(clonePath, "/tmp")
	ioutil.WriteFile("/tmp/e2e_log.txt", []byte(logFileE2E), 0644)

	// Parse the file to generate condepaths and add the corresponding results to graph.
	// TODO: Map service name from git path back to name
	gremlin.AddRuntimePathsToGraph("machine-config-operator",
		branchDetails[0].Revision, runtimelogs.CreateRuntimePaths("/tmp/e2e_log.txt"))

	touchPoints := serviceparser.GetTouchPointsOfPR(hunks, branchDetails)
	response := gremlin.GetTouchPointCoverage(touchPoints)

	render.JSON(w, r, response) // Return the same thing for now.
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
