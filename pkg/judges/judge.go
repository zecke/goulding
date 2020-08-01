package judges

import (
	"log"
	"sync"

	pb "github.com/zecke/goulding/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type sample struct {
	sample  float64
	comment string
}

type Results struct {
	mux     sync.Mutex
	samples map[string][]sample
}

type Judge interface {
	Judge(*Results) *pb.CanaryVerdict
}

func NewResults() *Results {
	return &Results{samples: make(map[string][]sample)}
}

func NewJudgeFromConfig(cfg *pb.CanaryConfig) (Judge, error) {
	if cfg.GetPrimitiveJudge() != nil {
		return NewPrimitiveJudge(cfg), nil
	}
	return nil, status.Error(codes.Unimplemented, "Unknown judge")
}

func (r *Results) Averages() map[string]float64 {
	r.mux.Lock()
	defer r.mux.Unlock()

	res := make(map[string]float64)

	for key, values := range r.samples {
		sum := 0.0
		for _, v := range values {
			sum += v.sample
		}
		res[key] = sum / float64(len(values))
	}

	return res
}

func (r *Results) RecordError(source, reason string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	log.Printf("Recording error=%s for source=%s\n", reason, source)

	item := r.samples[source]
	item = append(item, sample{0.0, reason})
	r.samples[source] = item
}

func (r *Results) RecordSuccess(source, reason string) {
	r.mux.Lock()
	defer r.mux.Unlock()

	log.Printf("Recording succes=%s for source=%s\n", reason, source)

	item := r.samples[source]
	item = append(item, sample{1.0, reason})
	r.samples[source] = item
}

func (r *Results) RecordValue(source string, value float64) {
	r.mux.Lock()
	defer r.mux.Unlock()

	log.Printf("Recording value=%v for source=%s\n", value, source)

	item := r.samples[source]
	item = append(item, sample{value, ""})
	r.samples[source] = item
}
