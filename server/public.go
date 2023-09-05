package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// public is the handler for any public documents, for example, the style sheet or images
func (web *Web) public(w http.ResponseWriter, r *http.Request) {
	if verbose {
		fmt.Println("Public called")
	}
	vars := mux.Vars(r)
	http.ServeFile(w, r, "public/"+vars["id"])
}
