package metcoll

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	pb "github.com/ArtemShalinFe/metcoll/proto/v1"
)

type GRPCClient struct {
	sl       *zap.SugaredLogger
	host     string
	clientIP string
	hashkey  []byte
}

func NewGRPCClient(cfg *configuration.ConfigAgent, sl *zap.SugaredLogger) (*GRPCClient, error) {
	c := &GRPCClient{
		host:     cfg.Server,
		clientIP: localIP(),
		hashkey:  cfg.Key,
		sl:       sl,
	}

	return c, nil
}

func (c *GRPCClient) BatchUpdateMetric(ctx context.Context, mcs <-chan []*metrics.Metrics, result chan<- error) {
	creds, err := credentials.NewClientTLSFromFile(
		"/Users/artem/Yandex.Disk.localized/Учеба/go-разработчик/metcoll/keys/cert.pem",
		"localhost")
	if err != nil {
		result <- fmt.Errorf("failed to load credentials: %v", err)
		return
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	retryopts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(2 * time.Second)),
		grpc_retry.WithCodes(grpc_retry.DefaultRetriableCodes...),
		grpc_retry.WithMax(3),
	}

	chain := grpc.WithChainUnaryInterceptor(
		c.clientCompressInterceptor,
		grpc_retry.UnaryClientInterceptor(retryopts...),
	)

	opts = append(opts, chain)

	conn, err := grpc.Dial(c.host, opts...)
	if err != nil {
		result <- fmt.Errorf("server is not available at %s, err: %w", c.host, err)
		return
	}
	defer conn.Close()
	mc := pb.NewMetcollClient(conn)

	headers := map[string]string{
		"X-Real-IP": c.clientIP,
		HashSHA256:  "",
	}

	for m := range mcs {
		var request pb.BatchUpdateRequest

		for _, mtrs := range m {
			pbm := c.convertMetric(mtrs)
			request.Metrics = append(request.Metrics, pbm)
		}

		if len(c.hashkey) != 0 {
			b, err := convertToBytes(request.Metrics)
			if err != nil {
				result <- fmt.Errorf("unable to convert metrics to bytes, err: %w", err)
				return
			}
			h := hmac.New(sha256.New, c.hashkey)

			h.Write(b)
			headers[HashSHA256] = hashBytesToString(h, nil)
		}

		mctx := metadata.NewOutgoingContext(ctx, metadata.New(headers))
		_, err := mc.Updates(mctx, &request)
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

func (c *GRPCClient) clientCompressInterceptor(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	opts = append(opts, grpc.UseCompressor(gzip.Name))

	err := invoker(ctx, method, req, reply, cc, opts...)

	if err != nil {
		return fmt.Errorf("compress %s was failed, err: %w", method, err)
	}
	return nil
}
