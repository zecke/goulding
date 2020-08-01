package sources

import (
	"context"
	"log"
	"net/http"

	"github.com/zecke/goulding/pkg/judges"
	pb "github.com/zecke/goulding/proto"
)

type HttpHealthSource struct {
	BaseSource

	url string
}

func NewHttpHealthSource(src *pb.Source, res *judges.Results, req *pb.CanaryRequest) *HttpHealthSource {
	p := HttpHealthSource{
		BaseSource: BaseSource{
			name: src.GetSourceName(),
			res:  res,
			req:  req,
		},
		url: expandTemplate(src.GetHealthSource().GetUrl(), req),
	}
	return &p
}

func (p *HttpHealthSource) Run(ctx context.Context) {
	log.Printf("Running http health check source=%s url=%s", p.name, p.url)

	req, err := http.NewRequestWithContext(ctx, "GET", p.url, nil)
	if err != nil {
		p.res.RecordError(p.name, err.Error())
		return
	}

	res, err := http.DefaultClient.Do(req)
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

	if res.StatusCode < 200 || res.StatusCode > 299 {
		p.res.RecordError(p.name, res.Status)
		return
	}

	p.res.RecordSuccess(p.name, res.Status)
}
