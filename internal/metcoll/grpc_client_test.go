package metcoll

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func NewSrvListener(srv MetcollServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()
	RegisterMetcollServer(server, srv)

	go func() {
		if err := server.Serve(listener); err != nil {
			zap.S().Errorf("grpc serve failed, err: %v", err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestGRPCClient_BatchUpdateMetric(t *testing.T) {
	ctx := context.Background()
	logger := zap.S()

	ctrl := gomock.NewController(t)
	MockServer := NewMockMetcollServer(ctrl)

	MockServer.EXPECT().Updates(gomock.Any(), gomock.Any()).Return(&BatchUpdateResponse{}, nil)

	cfg := &configuration.ConfigAgent{}
	cfg.Key = []byte("secretKeyHash")

	c, err := NewGRPCClient(ctx, cfg, logger)
	if err != nil {
		t.Errorf("an occured error when getting grpc client, err: %v", err)
	}

	opts := c.getDialOpts()
	creds := insecure.NewCredentials()
	opts = append(opts, grpc.WithTransportCredentials(creds))

	lis := NewSrvListener(MockServer)
	opts = append(opts, grpc.WithContextDialer(lis))

	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c.cc = conn

	t.Run("batch update metrics", func(t *testing.T) {
		mcs := make(chan []*metrics.Metrics, 1)
		var result chan error

		go func(mcs chan<- []*metrics.Metrics) {
			var ms1 []*metrics.Metrics
			ms1 = append(ms1, metrics.NewCounterMetric("counter1", int64(1)))
			ms1 = append(ms1, metrics.NewGaugeMetric("gauge1", float64(0.1)))

			select {
			case <-ctx.Done():
				return
			case mcs <- ms1:
			default:
			}
		}(mcs)

		tctx, cancel := context.WithTimeout(ctx, time.Duration(2*time.Second))
		defer cancel()
		go c.BatchUpdateMetric(tctx, mcs, result)

		go func(t *testing.T, result chan error) {
			for err := range result {
				t.Errorf("batch update metrics was failed, err: %v", err)

				select {
				case <-tctx.Done():
				default:
				}
			}
		}(t, result)

		<-tctx.Done()
	})
}
