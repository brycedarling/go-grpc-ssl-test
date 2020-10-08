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
	"google.golang.org/grpc/credentials"
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

func httpGrpcRouter(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
		if r.ProtoMajor == 2 { // && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	log.Println("Echo Service starting...")

	m := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist("brycedarling.com", "go-grpc-ssl-test-oc3j2.ondigitalocean.app", "localhost"),
		Prompt:     autocert.AcceptTOS,
	}
	creds := credentials.NewTLS(m.TLSConfig())
	opts := []grpc.ServerOption{grpc.Creds(creds)}

	grpcServer := grpc.NewServer(opts...)
	echopb.RegisterEchoServiceServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, world"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "https"
	}

	/*
		go func() {
			log.Printf("Serving on port %s...", port)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), m.HTTPHandler(nil)))
		}()
	*/

	go func() {
		httpServer := &http.Server{
			// Addr:      ":https",
			Addr:      fmt.Sprintf(":%s", port),
			Handler:   httpGrpcRouter(grpcServer, m.HTTPHandler(handler)),
			TLSConfig: m.TLSConfig(),
		}
		log.Printf("Serving on port %s...", port)
		log.Fatal(httpServer.ListenAndServeTLS("", ""))
	}()

	// Wait for ctrl-c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	// Block until a signal is received
	<-ch

	log.Println("Stopping the server...")
	grpcServer.Stop()
	log.Println("Stopped.")
}
