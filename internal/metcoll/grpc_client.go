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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type GRPCClient struct {
	cc       grpc.ClientConnInterface
	sl       *zap.SugaredLogger
	host     string
	clientIP string
	certPath string
	hashkey  []byte
}

func NewGRPCClient(ctx context.Context, cfg *configuration.ConfigAgent, sl *zap.SugaredLogger) (*GRPCClient, error) {
	clientIP, err := localIP()
	if err != nil {
		return nil, fmt.Errorf("an occured error when grpc agent getting local IP, err: %w", err)
	}

	c := &GRPCClient{
		host:     cfg.Server,
		clientIP: clientIP,
		hashkey:  cfg.Key,
		sl:       sl,
		certPath: cfg.CertFilePath,
	}

	return c, nil
}

func getClientCreds(certFilePath string) (credentials.TransportCredentials, error) {
	if certFilePath == "" {
		creds := insecure.NewCredentials()
		return creds, nil
	}

	creds, err := credentials.NewClientTLSFromFile(
		certFilePath,
		"")
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	return creds, nil
}

func (c *GRPCClient) setupConn(ctx context.Context) error {
	opts := c.getDialOpts()

	creds, err := getClientCreds(c.certPath)
	if err != nil {
		return fmt.Errorf("an occured error when getting client credentials: %w", err)
	}

	opts = append(opts, grpc.WithTransportCredentials(creds))
	conn, err := grpc.DialContext(ctx, c.host, opts...)
	if err != nil {
		return fmt.Errorf("server is not available at %s, err: %w", c.host, err)
	}

	c.cc = conn

	return nil
}

const defaultBackoffLinear = 2
const defaultMaxAttempt = 3

func (c *GRPCClient) getDialOpts() []grpc.DialOption {
	var opts []grpc.DialOption

	retryopts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(defaultBackoffLinear * time.Second)),
		grpc_retry.WithCodes(grpc_retry.DefaultRetriableCodes...),
		grpc_retry.WithMax(defaultMaxAttempt),
	}

	chain := grpc.WithChainUnaryInterceptor(
		c.clientCompressInterceptor,
		grpc_retry.UnaryClientInterceptor(retryopts...),
	)

	opts = append(opts, chain)

	return opts
}

func (c *GRPCClient) BatchUpdateMetric(ctx context.Context, mcs <-chan []*metrics.Metrics, result chan<- error) {
	mc := NewMetcollClient(c.cc)

	headers := map[string]string{
		realIP:     c.clientIP,
		HashSHA256: "",
	}

	for m := range mcs {
		var request BatchUpdateRequest

		for _, mtrs := range m {
			pbm := convertPBMetric(mtrs)
			request.Metrics = append(request.Metrics, pbm)
		}

		if len(c.hashkey) != 0 {
			b, err := convertToBytes(request.Metrics)
			if err != nil {
				result <- fmt.Errorf("unable to convert batch metrics to bytes, err: %w", err)
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

func convertPBMetric(m *metrics.Metrics) *Metric {
	var mt Metric
	mt.Id = m.ID

	switch m.MType {
	case metrics.CounterMetric:
		mt.Type = Metric_COUNTER
		mt.Delta = *m.Delta
	case metrics.GaugeMetric:
		mt.Type = Metric_GAUGE
		mt.Value = *m.Value
	default:
		mt.Type = Metric_UNKNOWN
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
