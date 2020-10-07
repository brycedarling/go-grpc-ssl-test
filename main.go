package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/brycedarling/go-grpc-ssl-test/internal/echopb"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	echopb.EchoServiceServer
}

func (*server) Echo(ctx context.Context, req *echopb.EchoRequest) (*echopb.EchoResponse, error) {
	message := req.GetMessage()
	log.Printf("Echo invoked: %v\n", message)
	return &echopb.EchoResponse{Message: message}, nil
}

func main() {
	log.Println("Echo Service starting...")

	s := grpc.NewServer()
	echopb.RegisterEchoServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		m := &autocert.Manager{
			Cache:      autocert.DirCache("certs"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("brycedarling.com"),
		}
		port := os.Getenv("PORT")
		if port == "" {
			port = "https"
		}
		s := &http.Server{
			Addr:      fmt.Sprintf(":%s", port),
			TLSConfig: m.TLSConfig(),
		}
		log.Printf("Serving on port %s...", port)
		log.Fatal(s.ListenAndServeTLS("", ""))
	}()

	// Wait for ctrl-c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	// Block until a signal is received
	<-ch

	log.Println("Stopping the server...")
	s.Stop()
	log.Println("Stopped.")
}
