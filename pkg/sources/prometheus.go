package sources

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/zecke/goulding/pkg/judges"
	pb "github.com/zecke/goulding/proto"
)

type PrometheusSource struct {
	BaseSource

	query, url string
}

func NewPrometheusSource(src *pb.Source, res *judges.Results, req *pb.CanaryRequest) *PrometheusSource {
	p := PrometheusSource{
		BaseSource: BaseSource{
			name: src.GetSourceName(),
			res:  res,
			req:  req,
		},
		query: expandTemplate(src.GetPrometheusSource().GetQuery(), req),
		url:   expandTemplate(src.GetPrometheusSource().GetServer(), req),
	}
	return &p
}

func (p *PrometheusSource) Run(ctx context.Context) {
	log.Printf("Running prometheus source=%s url=%s", p.name, p.url)

	client, err := api.NewClient(api.Config{
		Address: p.url,
	})
	if err != nil {
		p.res.RecordError(p.name, err.Error())
		return
	}

	v1api := v1.NewAPI(client)
	result, warnings, err := v1api.Query(ctx, p.query, time.Now())
	if err != nil {
		select {
		case <-ctx.Done():
			// Don't report if our time is up
			log.Printf("Request on cancelled context. source=%s Skipping: %v\n", p.name, err)
			return
		default:
			p.res.RecordError(p.name, err.Error())
			return
		}
	}
	if len(warnings) > 0 {
		log.Printf("Query warnings source=%s: %v\n", p.name, warnings)
	}
	switch result.Type() {
	case model.ValVector:
		v := result.(model.Vector)
		for _, sample := range v {
			p.res.RecordValue(p.name, float64(sample.Value))
		}
		return
	default:
		p.res.RecordError(p.name, fmt.Sprintf("result type=%v not implemented", result.Type()))
		return
	}
}
