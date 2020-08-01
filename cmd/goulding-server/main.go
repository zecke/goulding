package main

import (
	"flag"
	"log"
	"net"

	"github.com/zecke/goulding/pkg/config"
	"github.com/zecke/goulding/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	flgPort   = flag.String("grpc.listen", ":8343", "Port to listen on")
	flgConfig = flag.String("canaries.file", "examples/config/canaries.proto.txt", "Specify the config file for canaries")
)

func main() {
	flag.Parse()

	cfg, err := config.ParseConfig(*flgConfig)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	lis, err := net.Listen("tcp", *flgPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	server.NewServer(cfg).Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to run server: %v\n", err)
	}
}
