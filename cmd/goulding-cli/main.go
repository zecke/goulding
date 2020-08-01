package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/zecke/goulding/pkg/config"
	"github.com/zecke/goulding/pkg/server"
	pb "github.com/zecke/goulding/proto"
)

var (
	flgConfig  = flag.String("canaries.file", "examples/config/canaries.proto.txt", "Specify the config file for canaries")
	flgRequest = flag.String("canaries.request", "examples/request/request.proto.txt", "Make a single request and wait for completion.")
)

func parseRequest(fileName string) (*pb.CanaryRequest, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var req pb.CanaryRequest
	if err := proto.UnmarshalText(string(data), &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func main() {
	flag.Parse()

	request, err := parseRequest(*flgRequest)
	if err != nil {
		log.Fatalf("Failed to parse request: %v\n", err)
	}

	cfg, err := config.ParseConfig(*flgConfig)
	if err != nil {
		log.Fatalf("Failed to parse config: %v\n", err)

	}
	fmt.Printf("Parsed\n Config: %+v\n Request: %+v\n", cfg, request)

	s := server.NewServer(cfg)
	verdict, err := s.RunCanary(context.TODO(), request)
	if err != nil {
		log.Fatalf("Failed to canary: %v\n", err)
	}
	fmt.Printf("%+v\n", verdict)
}
