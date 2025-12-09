package main

import (
	"fmt"
	"net/http"
)

const PORT = ":8080"

func main() {
	mux := http.NewServeMux()

	server := NewServer()

	mux.HandleFunc("/put", server.handlePut)
	mux.HandleFunc("/get", server.handleGet)

	fmt.Println("Server is running on port: ", PORT)

	if err := http.ListenAndServe(PORT, mux); err != nil {
		fmt.Println("Server Error: ", err)
	}
}
