package config

import (
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	pb "github.com/zecke/goulding/proto"
)

func ParseConfig(fileName string) (*pb.CanaryConfigs, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfgs pb.CanaryConfigs
	if err := proto.UnmarshalText(string(data), &cfgs); err != nil {
		return nil, err
	}
	return &cfgs, err
}
