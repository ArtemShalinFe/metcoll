package main

import (
	"net/http"

	h "github.com/ArtemShalinFe/metcoll/internal/handlers"
)

func main() {

	mux := http.NewServeMux()
	mux.Handle(`/update/`, http.HandlerFunc(h.Update))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}

}
