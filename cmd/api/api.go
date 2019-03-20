package main

import (
	"encoding/json"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/ghpr"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

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
	prID int `json:"pr_id"`
	repoURL string `json:"repo_url"`
}

func RoutesPR() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", RunPRCoverage)
	return router
}

func RunPRCoverage(w http.ResponseWriter, r *http.Request) {
	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var msg PRPayload
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	ghpr.GetPRPayload(msg.repoURL, msg.prID, "/tmp")
	render.JSON(w, r, msg) // Return the same thing for now.
}

func main() {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
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
