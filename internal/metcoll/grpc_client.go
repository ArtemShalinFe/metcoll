package metcoll

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	pb "github.com/ArtemShalinFe/metcoll/proto/v1"
)

type GRPCClient struct {
	host string
}

func NewGRPCClient(cfg *configuration.ConfigAgent, sl *zap.SugaredLogger) (*GRPCClient, error) {
	c := &GRPCClient{
		host: cfg.Server,
	}

	return c, nil
}

func (c *GRPCClient) BatchUpdateMetric(ctx context.Context, mcs <-chan []*metrics.Metrics, result chan<- error) {
	conn, err := grpc.Dial(c.host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		result <- fmt.Errorf("server is not available at %s, err: %w", c.host, err)
		return
	}
	defer conn.Close()
	mc := pb.NewMetcollClient(conn)

	for m := range mcs {
		var request pb.BatchUpdateRequest

		for _, mtrs := range m {
			pbm := c.convertMetric(mtrs)
			request.Metrics = append(request.Metrics, pbm)
		}

		_, err := mc.Updates(ctx, &request)
		if err != nil {
			result <- err
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (c *GRPCClient) convertMetric(m *metrics.Metrics) *pb.Metric {
	var mt pb.Metric
	mt.ID = m.ID

	switch m.MType {
	case metrics.CounterMetric:
		mt.MType = pb.Metric_COUNTER
		mt.Delta = *m.Delta
	case metrics.GaugeMetric:
		mt.MType = pb.Metric_GAUGE
		mt.Value = *m.Value
	default:
		mt.MType = pb.Metric_UNKNOWN
	}

	return &mt
}
