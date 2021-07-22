package api

import "net/http"

type ServiceHandler struct {
}

func NewHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/msg", sendMessage)
	mux.HandleFunc("/", homePage)
	return mux
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	
}
func homePage(w http.ResponseWriter, r *http.Request) {

}