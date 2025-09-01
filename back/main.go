package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Define a handler function for the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hola Mundo")
	})

	// Start the server on port 8080
	port := ":8080"
	fmt.Printf("Servidor escuchando en http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
