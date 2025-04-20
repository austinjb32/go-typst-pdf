package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/generate", GenerateHandler).Methods("POST")
	router.HandleFunc("/template/list", ListTemplatesHandler).Methods("GET")

	router.HandleFunc("/template/new", UploadFormHandler).Methods("GET")
	router.HandleFunc("/template/upload", UploadTemplateHandler).Methods("POST")
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	return router
}
