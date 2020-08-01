package sources

import (
	"bytes"
	"context"
	"errors"
	"log"
	"text/template"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/zecke/goulding/pkg/judges"
	pb "github.com/zecke/goulding/proto"
)

type Source interface {
	// Run the source and fill results into the result set.
	Run(context.Context)
}

type BaseSource struct {
	name string
	res  *judges.Results
	req  *pb.CanaryRequest
}

type PeriodicSource struct {
	interval time.Duration
	source   Source
}

// NewSourceFromConfig creates a matching Source from the protobuf configuration
func NewSourceFromConfig(src *pb.Source, res *judges.Results, req *pb.CanaryRequest) (Source, error) {
	var source Source
	if src.GetPrometheusSource() != nil {
		source = NewPrometheusSource(src, res, req)
	} else if src.GetHealthSource() != nil {
		source = NewHttpHealthSource(src, res, req)
	} else {
		return nil, errors.New("No known source")
	}

	if src.GetExecutionMode() == pb.Source_MODE_PERIODIC {
		interval, err := ptypes.Duration(src.GetInterval())
		if err != nil {
			return nil, errors.New("Invalid periodic source interval")
		}
		return &PeriodicSource{interval: interval, source: source}, nil

	}
	return source, nil

}

// Run runs the underlying source periodically.
func (p *PeriodicSource) Run(ctx context.Context) {
	timer := time.NewTicker(p.interval)
	defer timer.Stop()

	p.source.Run(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			p.source.Run(ctx)
		}
	}
}

func expandTemplate(tmpl string, req *pb.CanaryRequest) string {
	t, err := template.New("sources").Parse(tmpl)
	if err != nil {
		log.Printf("Failed to parse template: %v\n", err)
		return tmpl
	}

	str := bytes.NewBuffer(nil)
	t.Execute(str, req.GetParameters())
	return str.String()

}
