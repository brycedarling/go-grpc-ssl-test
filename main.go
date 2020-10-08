package main

import (
	"context"
	"fmt"
	"log"
	"net"
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

func grpcHandler(s *grpc.Server, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
		log.Printf("ProtoMajor: %d", r.ProtoMajor)
		if r.ProtoMajor == 2 { // && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			s.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func redirectTLS() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.Host)
		u := r.URL
		u.Host = net.JoinHostPort(host, "443")
		u.Scheme = "https"
		http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
	})
}

func httpsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS user from IP: %s\n\nYour config is: %+v", r.RemoteAddr, r.TLS)
	})
}

func main() {
	log.Println("Echo Service starting...")

	host := os.Getenv("HOST")
	if host == "" {
		host = "brycedarling.com"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	m := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(host),
		Prompt:     autocert.AcceptTOS,
	}
	creds := credentials.NewTLS(m.TLSConfig())
	opts := []grpc.ServerOption{grpc.Creds(creds)}

	s := grpc.NewServer(opts...)
	log.Println("Starting gRPC services...")
	echopb.RegisterEchoServiceServer(s, &server{})
	reflection.Register(s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Println("Serving on port http...")
		log.Fatal(http.ListenAndServe(":http", redirectTLS()))
	}()

	go func() {
		log.Println("Serving on port https...")
		log.Fatal(http.Serve(autocert.NewListener(host), httpsHandler()))
	}()

	go func() {
		log.Printf("Serving on port %s...", port)
		log.Fatal(s.Serve(lis))
	}()

	// Wait for ctrl-c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	// Block until a signal is received
	<-ch

	log.Println("Stopping the server...")
	s.Stop()
	lis.Close()
	log.Println("Stopped.")
}
