package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	fmt.Printf("ZarishSphere service starting on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
