package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("OpenBPL API Starting...")

	http.HandleFunc("/health", healthCheck)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is healthy!"))
}
