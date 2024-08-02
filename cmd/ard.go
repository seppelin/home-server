package main

import (
	"fmt"
	"io"
	"net/http"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got PUT / request\n")
	b, _ := io.ReadAll(r.Body)
	fmt.Println(string(b))
	fmt.Fprintf(w, "Received!")
}

func main() {
	http.HandleFunc("PUT /", handleRoot)
	_ = http.ListenAndServe(":2727", nil)
}
