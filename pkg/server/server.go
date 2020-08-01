package server

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/zecke/goulding/pkg/judges"
	"github.com/zecke/goulding/pkg/sources"
	pb "github.com/zecke/goulding/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	configs *pb.CanaryConfigs
}

func NewServer(configs *pb.CanaryConfigs) *server {
	s := &server{configs: configs}
	return s
}

func (s *server) Register(grpcServer *grpc.Server) {
	pb.RegisterCanaryServiceServer(grpcServer, s)
}

// runSource will execute a single source and add one or more results
func (s *server) runSource(ctx context.Context, src *pb.Source, res *judges.Results, req *pb.CanaryRequest) {
	log.Printf("Running source name='%s' for canary='%s' mode=%s\n", src.GetSourceName(), req.GetCanary(), src.GetExecutionMode())

	source, err := sources.NewSourceFromConfig(src, res, req)
	if err != nil {
		res.RecordError(src.GetSourceName(), "Failed to create source")
		return
	}

	source.Run(ctx)
}

// runAction will execute an action _after_ canarying is complete.
func (s *server) runAction(action *pb.Action, req *pb.CanaryRequest) {
	// TODO(zecke): Add retry and back-off...
	// TODO(zecke): Really run actions
}

// doCanary runs the canary by executing the sources and presenting the result to a judge.
func (s *server) doCanary(ctx context.Context, cfg *pb.CanaryConfig, req *pb.CanaryRequest) (*pb.CanaryVerdict, error) {
	var wg sync.WaitGroup
	res := judges.NewResults()

	canaryingTime, err := ptypes.Duration(cfg.GetCanaryingTime())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Can not convert canarying time")
	}

	srcCtx, cancel := context.WithTimeout(ctx, canaryingTime)
	defer cancel()

	for _, source := range cfg.GetSources() {
		if source.GetExecutionMode() == pb.Source_MODE_PERIODIC {
			wg.Add(1)
			go func(src *pb.Source) {
				defer wg.Done()
				s.runSource(srcCtx, src, res, req)
			}(source)
		}
	}

	select {
	case <-ctx.Done():
		return nil, status.Errorf(codes.Canceled, "Cancelled the canary request")
	case <-time.After(canaryingTime):
		// nothing and continue
	}

	for _, source := range cfg.GetSources() {
		if source.GetExecutionMode() == pb.Source_MODE_SINGLESHOT {
			wg.Add(1)
			go func(src *pb.Source) {
				defer wg.Done()
				s.runSource(ctx, src, res, req)
			}(source)
		}
	}

	// TODO(zecke): This expects all sources are using the ctx correctly.
	wg.Wait()

	judge, err := judges.NewJudgeFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	verdict := judge.Judge(res)
	var actions []*pb.Action
	if verdict.Pass {
		actions = cfg.GetPassActions()
	} else {
		actions = cfg.GetFailActions()
	}

	for _, action := range actions {
		go s.runAction(action, req)
	}

	return verdict, nil
}

func (s *server) RunCanary(ctx context.Context, req *pb.CanaryRequest) (*pb.CanaryVerdict, error) {
	log.Printf("Requested to canary '%s'\n", req.GetCanary())
	for _, cfg := range s.configs.GetCanaryConfig() {
		if cfg.GetCanaryName() == req.GetCanary() {
			return s.doCanary(ctx, cfg, req)
		}
	}

	return nil, status.Errorf(codes.NotFound, "Canary with name='%s' not configured.", req.GetCanary())
}
