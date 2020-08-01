package judges

import (
	"fmt"

	pb "github.com/zecke/goulding/proto"
)

type PrimitiveJudge struct {
	cfg *pb.CanaryConfig
}

func NewPrimitiveJudge(cfg *pb.CanaryConfig) *PrimitiveJudge {
	return &PrimitiveJudge{cfg: cfg}
}

func (j *PrimitiveJudge) Judge(res *Results) *pb.CanaryVerdict {
	// TODO(freyth): Make sure that every source provided a result.
	threshold := j.cfg.GetPrimitiveJudge().GetPerSourceSuccessThreshold()
	avgs := res.Averages()

	for src, avg := range avgs {
		if avg < threshold {
			return &pb.CanaryVerdict{Pass: false, Result: fmt.Sprintf("source=%v value=%v below threshold=%v", src, avg, threshold)}
		}
	}

	return &pb.CanaryVerdict{Pass: true, Result: "LGTM"}
}
