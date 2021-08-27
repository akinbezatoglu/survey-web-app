package api

import (
	"net/http"

	"github.com/akinbezatoglu/survey-web-app/pkg/api/response"
	"github.com/akinbezatoglu/survey-web-app/pkg/database/mongodb"

	"github.com/gorilla/mux"
)

// Config represents the API configuration
type Config struct {
	Domain        string `yaml:"domain"`
	SigningSecret string `yaml:"signing_secret"`
}

// API represents the structure of the API
type API struct {
	Router *mux.Router

	config *Config
	db     mongodb.DB
}

// New returns the api settings
func New(config *Config, db mongodb.DB, router *mux.Router) (*API, error) {
	api := &API{
		config: config,
		db:     db,
		Router: router,
	}

	// Endpoint for browser preflight requests
	api.Router.Methods("OPTIONS").HandlerFunc(api.corsMiddleware(api.preflightHandler))

	// Endpoint for healtcheck
	api.Router.HandleFunc("/api/v1/health", api.corsMiddleware(api.logMiddleware(api.healthHandler))).Methods("GET")

	// User
	api.Router.HandleFunc("/api/v1/auth", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.UserValidate)))).Methods("GET")
	api.Router.HandleFunc("/api/v1/auth/login", api.corsMiddleware(api.logMiddleware(api.userLoginPostHandler))).Methods("POST")
	api.Router.HandleFunc("/api/v1/auth/signup", api.corsMiddleware(api.logMiddleware(api.userSignupPostHandler))).Methods("POST")
	api.Router.HandleFunc("/api/v1/profile", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.userProfileGetHandler)))).Methods("GET")
	api.Router.HandleFunc("/api/v1/profile/{userid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.userProfileUpdateHandler)))).Methods("PUT")

	// Home
	api.Router.HandleFunc("/api/v1/f/{userid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.GetAllFormsHandler)))).Methods("GET")
	api.Router.HandleFunc("/api/v1/f/{userid}/less", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.GetAllFormsWithLessValueHandler)))).Methods("GET")
	api.Router.HandleFunc("/api/v1/f/{userid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.CreateFormHandler)))).Methods("POST")

	// Form
	api.Router.HandleFunc("/api/v1/f/{formid}/view", api.corsMiddleware(api.logMiddleware(api.formViewGetHandler))).Methods("GET")
	api.Router.HandleFunc("/api/v1/f/{formid}/view", api.corsMiddleware(api.logMiddleware(api.formViewPostHandler))).Methods("POST")
	api.Router.HandleFunc("/api/v1/f/{userid}/{formid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.formGetHandler)))).Methods("GET")
	api.Router.HandleFunc("/api/v1/f/{userid}/{formid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.formPutHandler)))).Methods("PUT")
	api.Router.HandleFunc("/api/v1/f/{userid}/{formid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.formDeleteHandler)))).Methods("DELETE")

	api.Router.HandleFunc("/api/v1/f/{userid}/{formid}/response", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.formResponseGetHandler)))).Methods("GET")

	api.Router.HandleFunc("/api/v1/q/{userid}/{formid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.questionPostHandler)))).Methods("POST")
	api.Router.HandleFunc("/api/v1/q/{userid}/{formid}/copy", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.questionPostCopyHandler)))).Methods("POST")
	api.Router.HandleFunc("/api/v1/q/{userid}/{formid}/{questionid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.questionPutHandler)))).Methods("PUT")
	api.Router.HandleFunc("/api/v1/q/{userid}/{formid}/{questionid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.questionDeleteHandler)))).Methods("DELETE")

	api.Router.HandleFunc("/api/v1/o/{userid}/{formid}/{questionid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.optionPostHandler)))).Methods("POST")
	api.Router.HandleFunc("/api/v1/o/{userid}/{formid}/{questionid}/{optionid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.optionPutHandler)))).Methods("PUT")
	api.Router.HandleFunc("/api/v1/o/{userid}/{formid}/{questionid}/{optionid}", api.corsMiddleware(api.logMiddleware(api.jwtMiddleware(api.optionDeleteHandler)))).Methods("DELETE")

	return api, nil
}

func (a *API) healthHandler(w http.ResponseWriter, r *http.Request) {
	response.Write(w, r, struct {
		Status string `json:"status"`
	}{
		"ok",
	})
}

func (a *API) preflightHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
