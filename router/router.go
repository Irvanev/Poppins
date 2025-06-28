package router

import (
	"net/http"
	"poppins/handlers"

	"github.com/gorilla/mux"
)

func NewRouter(uh *handlers.UserHandler, ah *handlers.AdHandler) *mux.Router {
	r := mux.NewRouter()

	// User endpoints
	r.HandleFunc("/users", uh.Create).Methods("POST")
	r.HandleFunc("/users/{telegramId}", uh.Get).Methods("GET")
	r.HandleFunc("/users/{telegramId}", uh.Delete).Methods("DELETE")

	// Список объявлений конкретного пользователя
	r.HandleFunc("/users/{telegramId}/ads", ah.ListByTelegram).Methods("GET")

	// Ad endpoints
	r.HandleFunc("/ads", ah.Create).Methods("POST")
	r.HandleFunc("/ads/{id}", ah.Get).Methods("GET")
	r.HandleFunc("/ads", ah.Search).Methods("GET")
	r.HandleFunc("/ads/{id}", ah.Update).Methods("PUT")
	r.HandleFunc("/ads/{id}", ah.Delete).Methods("DELETE")
	r.HandleFunc("/ads/{id}/archive", ah.Archive).Methods("PATCH")

	r.PathPrefix("/" + ah.Bucket + "/").
		Handler(http.StripPrefix("/"+ah.Bucket+"/", http.FileServer(http.Dir("."))))

	return r
}
