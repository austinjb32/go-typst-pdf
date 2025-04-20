package server

import (
	"context"
	"go-typst-pdf/pdf"
	"go-typst-pdf/proto"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	proto.UnimplementedPDFServiceServer
}

func (s *server) GeneratePDF(ctx context.Context, req *proto.PDFRequest) (*proto.PDFResponse, error) {
	data := make(map[string]interface{})
	for k, v := range req.Data {
		data[k] = v
	}
	url, err := pdf.GenerateAndUpload(req.Template, data)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return nil, err
	}
	return &proto.PDFResponse{Url: url}, nil
}

func StartGRPC(listener net.Listener) {
	log.Println("Starting gRPC server...")
	s := grpc.NewServer()
	proto.RegisterPDFServiceServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
