package metcoll

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
	pb "github.com/ArtemShalinFe/metcoll/proto/v1"
)

type MetricService struct {
	pb.UnimplementedMetcollServer
	storage Storage
	log     *zap.SugaredLogger
}

func NewMetricService(s Storage, sl *zap.SugaredLogger) *MetricService {
	return &MetricService{
		storage: s,
		log:     sl,
	}
}

func (ms *MetricService) Updates(ctx context.Context, request *pb.BatchUpdateRequest) (*pb.BatchUpdateResponse, error) {
	var response pb.BatchUpdateResponse

	mtrs := make([]*metrics.Metrics, len(request.GetMetrics()))

	for i := 0; i < len(request.GetMetrics()); i++ {
		m := request.Metrics[i]

		mtr, err := ms.convertMetric(m)
		if err != nil {
			response.Error = err.Error()
			return &response, nil
		}

		mtrs[i] = mtr
	}

	ums, errs, err := metrics.BatchUpdate(ctx, mtrs, ms.storage)
	if err != nil {
		ms.log.Errorf("BatchUpdate update error: %w", err)
		response.Error = "batch update metric was failed"
		return &response, nil
	}
	if len(errs) > 0 {
		for _, err := range errs {
			ms.log.Errorf("BatchUpdate update error: %w", err)
		}

		response.Error = "not all metrics have been updated"
		return &response, nil
	}

	for _, um := range ums {
		ms.log.Debugf("Metric %s was updated. New value: %s", strings.TrimSpace(um.ID), um.String())
	}

	return &response, nil
}

func (ms *MetricService) Update(ctx context.Context, request *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	var response pb.UpdateResponse

	mtr, err := ms.convertMetric(request.GetMetric())
	if err != nil {
		response.Error = err.Error()
		return &response, nil
	}

	ms.log.Infof("Trying update %s metric %s with value: %s", mtr.MType, mtr.ID, mtr.String())

	if err := mtr.Update(ctx, ms.storage); err != nil {
		if !errors.Is(err, storage.ErrNoRows) {
			ms.log.Errorf("an error occurred while updating the metric, err: %w", err)
			response.Error = "an error occurred while updating the metric"
			return &response, nil
		}
	}

	return &response, nil
}

func (ms *MetricService) convertMetric(pbm *pb.Metric) (*metrics.Metrics, error) {
	switch pbm.MType {
	case pb.Metric_COUNTER:
		return metrics.NewCounterMetric(pbm.GetID(), pbm.GetDelta()), nil
	case pb.Metric_GAUGE:
		return metrics.NewGaugeMetric(pbm.GetID(), pbm.GetValue()), nil
	default:
		t := "metric %s has unknow type: %s"
		ms.log.Infof(t, pbm.GetID(), pbm.GetMType())
		return nil, fmt.Errorf(t, pbm.GetID(), pbm.GetMType())
	}
}

type GRPCServer struct {
	grpcServer *grpc.Server
	addr       string
	ms         *MetricService
	sl         *zap.SugaredLogger
}

func NewGRPCServer(s Storage, cfg *configuration.Config, sl *zap.SugaredLogger) (*GRPCServer, error) {
	return &GRPCServer{
		grpcServer: grpc.NewServer(),
		addr:       cfg.Address,
		ms:         NewMetricService(s, sl),
		sl:         sl,
	}, nil
}

func (s *GRPCServer) ListenAndServe() error {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("an occured error when trying listen address %s, err: %w", s.addr, err)
	}
	pb.RegisterMetcollServer(s.grpcServer, s.ms)

	if err := s.grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("grpc server serve err: %w", err)
	}

	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.grpcServer.GracefulStop()
	return nil
}
