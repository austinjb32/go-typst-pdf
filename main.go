package main

import (
	"context"
	"go-typst-pdf/api"
	"go-typst-pdf/pdf"
	"go-typst-pdf/queue"
	"go-typst-pdf/server"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Start the HTTP API server
	router := api.SetupRouter()
	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Start the gRPC server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on :50051: %v", err)
	}
	grpcSrv := server.NewGRPCServer()
	go func() {
		log.Println("Starting gRPC server on :50051")
		if err := grpcSrv.Serve(listener); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	// Gracefully shut down HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Gracefully stop gRPC server
	grpcSrv.GracefulStop()

	// Close the job queue
	queue.CloseJobQueue()

	log.Println("Server stopped")
}
