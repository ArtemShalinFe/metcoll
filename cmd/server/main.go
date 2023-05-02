package main

import (
	"net/http"

	h "github.com/ArtemShalinFe/metcoll/internal/handlers"
)

func main() {

	r := h.ChiRouter()
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}

}
