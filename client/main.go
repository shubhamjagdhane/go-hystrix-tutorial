package main

import (
	"fmt"
	"net/http"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, Circuit-Breaker")
		fmt.Fprintln(w, "Hello, Circuit-Breaker")
	})

	http.ListenAndServe(":8081", nil)
}
