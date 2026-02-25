package main

import (
	"net/http"

	"github.com/taerc/entviz/examples/ent"
)

func main() {
	http.ListenAndServe("localhost:3002", ent.ServeEntviz())
}
