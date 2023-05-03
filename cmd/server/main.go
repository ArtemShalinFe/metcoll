package main

import (
	"flag"
	"fmt"
	"net/http"

	h "github.com/ArtemShalinFe/metcoll/internal/handlers"
)

func main() {

	serverEndPoint := flag.String("a", "localhost:8080", "server end point")

	flag.Parse()

	r := h.ChiRouter()

	fmt.Printf("server end point is %s", *serverEndPoint)

	err := http.ListenAndServe(*serverEndPoint, r)
	if err != nil {
		panic(err)
	}

}
