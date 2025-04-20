package main

import (
	"go-typst-pdf/api"
	"go-typst-pdf/pdf"
	"go-typst-pdf/queue"
	"go-typst-pdf/server"
	"log"
	"net"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	go queue.StartWorkerPool(5) // Adjust the number of workers as needed

	// Initialize the template cache
	pdf.InitTemplateCache()

	// Start the HTTP API server in a goroutine
	go func() {
		router := api.SetupRouter()
		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Start the gRPC server in a goroutine
	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on :50051: %v", err)
		}
		log.Println("Starting gRPC server on :50051")
		server.StartGRPC(listener)
	}()

	// Block the main goroutine to keep the application running
	select {}
}
