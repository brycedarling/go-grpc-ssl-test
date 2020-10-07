package main

import (
	"context"
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
		log.Println("Serving...")
		log.Fatal(http.Serve(autocert.NewListener("brycedarling.com"), s))
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
