package metcoll

import (
	"context"
	"errors"
	"net"
	reflect "reflect"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	pb "github.com/ArtemShalinFe/metcoll/proto/v1"
	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/test/bufconn"
)

type dialer struct {
	lis *bufconn.Listener
}

func NewDialer(t *testing.T, stg Storage) (*dialer, error) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	cfg := &configuration.Config{}

	s, err := NewGRPCServer(stg, cfg, zap.S())
	if err != nil {
		t.Fatalf("an occured error when initial grpc server, err: %v", err)
	}
	srv := NewMetricService(stg, zap.S())

	pb.RegisterMetcollServer(s, srv)
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	}()

	return &dialer{
		lis: lis,
	}, nil
}

func (d *dialer) bufDialer(context.Context, string) (net.Conn, error) {
	return d.lis.Dial()
}

func TestMetricService_Update(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockStorage(ctrl)
	stg.EXPECT().AddInt64Value(gomock.Any(), metricc, int64(11)).
		Return(int64(11), nil)

	stg.EXPECT().AddInt64Value(gomock.Any(), metricc, int64(11)).
		Return(int64(22), nil)

	stg.EXPECT().SetFloat64Value(gomock.Any(), metricg, float64(31.1)).
		Return(float64(31.1), nil)

	stg.EXPECT().SetFloat64Value(gomock.Any(), metricg, float64(32.1)).
		Return(float64(0), errors.New("unknow error"))

	d, err := NewDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	type args struct {
		metric *metrics.Metrics
	}
	tests := []struct {
		name         string
		args         args
		want         *pb.UpdateResponse
		wantPBMetric *pb.Metric
		wantErr      bool
	}{
		{
			name: "#1",
			args: args{
				metric: metrics.NewCounterMetric(metricc, 11),
			},
			want: &pb.UpdateResponse{},
			wantPBMetric: &pb.Metric{
				ID:    metricc,
				MType: pb.Metric_COUNTER,
				Delta: 11,
			},
			wantErr: false,
		},
		{
			name: "#2",
			args: args{
				metric: metrics.NewCounterMetric(metricc, 11),
			},
			want: &pb.UpdateResponse{},
			wantPBMetric: &pb.Metric{
				ID:    metricc,
				MType: pb.Metric_COUNTER,
				Delta: 22,
			},
			wantErr: false,
		},
		{
			name: "#3",
			args: args{
				metric: metrics.NewGaugeMetric(metricg, 31.1),
			},
			want:    &pb.UpdateResponse{},
			wantErr: false,
		},
		{
			name: "#4",
			args: args{
				metric: &metrics.Metrics{
					ID:    "someUnknowType",
					MType: "unknow",
				},
			},
			want:    &pb.UpdateResponse{},
			wantErr: true,
		},
		{
			name: "#5",
			args: args{
				metric: metrics.NewGaugeMetric(metricg, 32.1),
			},
			want:    &pb.UpdateResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(d.bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()
			client := pb.NewMetcollClient(conn)

			req := &pb.UpdateRequest{Metric: convertPBMetric(tt.args.metric)}
			got, err := client.Update(ctx, req)
			if err != nil && !tt.wantErr {
				t.Errorf("response MetcollClient.Update() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr && tt.wantPBMetric != nil {
				if !reflect.DeepEqual(got.Metric, tt.wantPBMetric) {
					t.Errorf("response Metric not equals, got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMetricService_Updates(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockStorage(ctrl)

	var ms []*metrics.Metrics
	ms = append(ms, metrics.NewCounterMetric(metricc, 1),
		metrics.NewGaugeMetric(metricg, 1.2))

	counters := make(map[string]int64)
	counters[metricc] = 1
	stg.EXPECT().BatchAddInt64Value(gomock.Any(), counters).Times(1).Return(counters, nil, nil)

	gauges := make(map[string]float64)
	gauges[metricg] = 1.2
	stg.EXPECT().BatchSetFloat64Value(gomock.Any(), gauges).Times(1).Return(gauges, nil, nil)

	var counterMetrics []*metrics.Metrics
	counterMetrics = append(counterMetrics, metrics.NewCounterMetric(metricc, 1))

	var gaugeMetrics []*metrics.Metrics
	gaugeMetrics = append(gaugeMetrics, metrics.NewGaugeMetric(metricg, 1.2))

	updatedCounterMetrics := make(map[string]int64)
	updatedCounterMetrics[metricc] = 1
	updatedGaugeMetrics := make(map[string]float64)
	updatedGaugeMetrics[metricg] = 1.2

	stg.EXPECT().BatchSetFloat64Value(gomock.Any(), updatedGaugeMetrics).Times(1).
		Return(nil, nil, errors.New("error batch update float64"))

	stg.EXPECT().BatchSetFloat64Value(gomock.Any(), make(map[string]float64)).Times(1).
		Return(make(map[string]float64), nil, nil)

	stg.EXPECT().BatchAddInt64Value(gomock.Any(), updatedCounterMetrics).Times(1).
		Return(nil, nil, errors.New("error batch update int64"))

	d, err := NewDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	type args struct {
		metrics []*metrics.Metrics
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.UpdateResponse
		wantErr bool
	}{
		{
			name: "#1",
			args: args{
				metrics: ms,
			},
			want:    &pb.UpdateResponse{},
			wantErr: false,
		},
		{
			name: "#2",
			args: args{
				metrics: gaugeMetrics,
			},
			want:    &pb.UpdateResponse{},
			wantErr: true,
		},
		{
			name: "#3",
			args: args{
				metrics: counterMetrics,
			},
			want:    &pb.UpdateResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(d.bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()
			client := pb.NewMetcollClient(conn)

			var req pb.BatchUpdateRequest
			for _, mtrs := range tt.args.metrics {
				pbm := convertPBMetric(mtrs)
				req.Metrics = append(req.Metrics, pbm)
			}

			got, err := client.Updates(ctx, &req)
			if err != nil && !tt.wantErr {
				t.Errorf("response MetcollClient.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}
