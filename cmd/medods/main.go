package main

import (
	"fmt"
	"log"
	"medods/internal/handlers"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/", handlers.AuthGuid)
	mux.HandleFunc("/refresh/", handlers.RefreshToken)

	server := http.Server{
		Addr:         ":4000",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	fmt.Println("HI!")
	fmt.Println(os.Getenv("DB_URL"))
	log.Fatal(server.ListenAndServe())
}
