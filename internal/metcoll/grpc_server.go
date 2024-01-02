package metcoll

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

type MetricService struct {
	UnimplementedMetcollServer
	storage Storage
	log     *zap.SugaredLogger
}

func NewMetricService(s Storage, sl *zap.SugaredLogger) *MetricService {
	return &MetricService{
		storage: s,
		log:     sl,
	}
}

func (ms *MetricService) Updates(ctx context.Context, request *BatchUpdateRequest) (*BatchUpdateResponse, error) {
	var response BatchUpdateResponse

	mtrs := make([]*metrics.Metrics, len(request.GetMetrics()))

	for i := 0; i < len(request.GetMetrics()); i++ {
		m := request.Metrics[i]

		mtr, err := convertMetric(m)
		if err != nil {
			response.Error = err.Error()
			return &response, fmt.Errorf("an error occured while convert pb metric, err: %w", err)
		}

		mtrs[i] = mtr
	}

	ums, errs, err := metrics.BatchUpdate(ctx, mtrs, ms.storage)
	if err != nil {
		ms.log.Errorf("BatchUpdate update error: %w", err)
		response.Error = "batch update metric was failed"
		return &response, fmt.Errorf("an error occured while convert pb batch metric, err: %w", err)
	}
	if len(errs) > 0 {
		for _, err := range errs {
			ms.log.Errorf("BatchUpdate update error: %w", err)
		}

		response.Error = "not all metrics have been updated"
		return &response, fmt.Errorf("an error occured while update metrics, err: %w", err)
	}

	for _, um := range ums {
		ms.log.Debugf("Metric %s was updated. New value: %s", strings.TrimSpace(um.ID), um.String())
	}

	return &response, nil
}

func (ms *MetricService) Update(ctx context.Context, request *UpdateRequest) (*UpdateResponse, error) {
	var response UpdateResponse

	mtr, err := convertMetric(request.GetMetric())
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

	mpb := convertPBMetric(mtr)
	response.Metric = mpb

	return &response, nil
}

func (ms *MetricService) ReadMetric(ctx context.Context, request *ReadMetricRequest) (*ReadMetricResponse, error) {
	var response ReadMetricResponse

	mtr, err := convertMetric(request.GetMetric())
	if err != nil {
		response.Error = err.Error()
		return &response, nil
	}

	if err := mtr.Get(ctx, ms.storage); err != nil {
		if !errors.Is(err, storage.ErrNoRows) {
			ms.log.Errorf("an error occurred while reading the metric, err: %w", err)
			response.Error = "an error occurred while reading the metric"
			return &response, nil
		}
	}

	mpb := convertPBMetric(mtr)
	response.Metric = mpb

	return &response, nil
}

func (ms *MetricService) MetricList(ctx context.Context, request *MetricListRequest) (*MetricListResponse, error) {
	var response MetricListResponse

	mts, err := ms.storage.GetDataList(ctx)
	if err != nil {
		ms.log.Errorf("an error occurred while getting metric list, err: %w", err)
		response.Error = "an error occurred while getting metric list"
		return &response, nil
	}

	list := ""
	for _, v := range mts {
		list += fmt.Sprintf(`<p>%s</p>`, v)
	}

	response.Htmlpage = fmt.Sprintf(templateMetricList(), list)
	return &response, nil
}

func convertMetric(pbm *Metric) (*metrics.Metrics, error) {
	switch pbm.Type {
	case Metric_COUNTER:
		return metrics.NewCounterMetric(pbm.GetId(), pbm.GetDelta()), nil
	case Metric_GAUGE:
		return metrics.NewGaugeMetric(pbm.GetId(), pbm.GetValue()), nil
	default:
		return nil, fmt.Errorf("metric %s has unknow type: %s", pbm.GetId(), pbm.GetType())
	}
}

type GRPCServer struct {
	grpcServer    *grpc.Server
	addr          string
	trustedSubnet *net.IPNet
	ms            *MetricService
	sl            *zap.SugaredLogger
	hashkey       []byte
}

func NewGRPCServer(s Storage, cfg *configuration.Config, sl *zap.SugaredLogger) (*GRPCServer, error) {
	srv := &GRPCServer{
		addr:          cfg.Address,
		ms:            NewMetricService(s, sl),
		sl:            sl,
		trustedSubnet: parseTrustedSubnet(cfg.TrustedSubnet),
		hashkey:       cfg.Key,
	}

	creds, err := getServerCreds(cfg)
	if err != nil {
		return nil, err
	}

	opt := grpc.ChainUnaryInterceptor(
		srv.requestLogger(),
		srv.resolverIP(),
		srv.hashChecker(),
	)
	srv.grpcServer = grpc.NewServer(grpc.Creds(creds), opt)

	return srv, nil
}

func (s *GRPCServer) RegisterService(desc *grpc.ServiceDesc, impl any) {
	s.grpcServer.RegisterService(desc, impl)
}

func (s *GRPCServer) Serve(lis net.Listener) error {
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("an occured error when server serve request, err: %v", err)
	}
	return nil
}

func (s *GRPCServer) ListenAndServe() error {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("an occured error when trying listen address %s, err: %w", s.addr, err)
	}
	RegisterMetcollServer(s.grpcServer, s.ms)

	if err := s.grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("an occured error when grpc server serve, err: %w", err)
	}

	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.grpcServer.GracefulStop()
	return nil
}

func (s *GRPCServer) requestLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		if err != nil {
			s.sl.Errorf("RPC request err: %v", err)
		} else {
			md, _ := metadata.FromIncomingContext(ctx)
			size, err := responseSize(resp)
			if err != nil {
				s.sl.Errorf("grpc response size calculate, err:%w", err)
			}
			s.sl.Infof("RPC request method: %s, header: %v, body: %s, duration: %s, responseSize: %d",
				info.FullMethod, md, req, duration, size,
			)
		}

		return resp, nil
	}
}

func (s *GRPCServer) resolverIP() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if s.trustedSubnet == nil {
			resp, err := handler(ctx, req)
			if err != nil {
				return nil, status.Errorf(codes.Unknown,
					"handler in resolver ip interceptor was failed, err: %v", err)
			}
			return resp, nil
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Aborted,
				"'X-Real-IP' header is required")
		}

		ips := md.Get("X-Real-IP")
		if len(ips) == 0 {
			return nil, status.Error(codes.Aborted,
				"'X-Real-IP' header not contain elemets")
		}

		ipStr := strings.TrimSpace(ips[0])
		if strings.TrimSpace(ipStr) == "" {
			return nil, status.Error(codes.Aborted,
				"first element in 'X-Real-IP' header is empty")
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, status.Error(codes.Aborted,
				"first element in 'X-Real-IP' header is not IP")
		}

		if !s.trustedSubnet.Contains(ip) {
			return nil, status.Error(codes.Aborted,
				"trusted network does not contain the first element in 'X-Real-IP' header")
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, status.Errorf(codes.Unknown,
				"handler in resolver ip interceptor was failed, err: %v", err)
		}
		return resp, nil
	}
}

func (s *GRPCServer) hashChecker() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if len(s.hashkey) == 0 {
			resp, err := handler(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("unable to check hash, err: %w", err)
			}
			return resp, nil
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Aborted,
				"'%s' header is required", HashSHA256)
		}

		hashes := md.Get(HashSHA256)
		if len(hashes) == 0 {
			return nil, status.Errorf(codes.Aborted,
				"request not contains values in header '%s'", HashSHA256)
		}

		hash := strings.TrimSpace(hashes[0])

		correctHash, err := s.correctRequestHash(req)
		if err != nil {
			return nil, status.Errorf(codes.Aborted,
				"an occured error when getting correct request hash, err: %v", err)
		}

		header := metadata.New(map[string]string{HashSHA256: correctHash})
		if err := grpc.SendHeader(ctx, header); err != nil {
			return nil, status.Errorf(codes.Internal, "unable to send '%s' header", HashSHA256)
		}

		if correctHash != hash {
			return nil, status.Error(codes.Aborted, "incorrect hash")
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("unable to convert metrics to bytes, err: %w", err)
		}
		return resp, nil
	}
}

func (s *GRPCServer) messageHash(message any) (string, error) {
	b, err := convertToBytes(message)
	if err != nil {
		return "", fmt.Errorf("unable to convert message to bytes, err: %w", err)
	}
	h := hmac.New(sha256.New, s.hashkey)
	h.Write(b)
	return hashBytesToString(h, nil), nil
}

func (s *GRPCServer) correctRequestHash(req any) (string, error) {
	switch r := req.(type) {
	case *BatchUpdateRequest:
		correctHash, err := s.messageHash(r.GetMetrics())
		if err != nil {
			return "", fmt.Errorf("batch update - bad request, err: %v", err)
		}
		return correctHash, nil
	case *UpdateRequest:
		correctHash, err := s.messageHash(r.GetMetric())
		if err != nil {
			return "", fmt.Errorf("update - bad request, err: %v", err)
		}
		return correctHash, nil
	case *ReadMetricRequest:
		correctHash, err := s.messageHash(r.GetMetric())
		if err != nil {
			return "", fmt.Errorf("read - bad request, err: %v", err)
		}
		return correctHash, nil
	default:
		return "", nil
	}
}

func getServerCreds(cfg *configuration.Config) (credentials.TransportCredentials, error) {
	if cfg.CertFilePath != "" && cfg.PrivateCryptoKey != "" {
		creds, err := credentials.NewServerTLSFromFile(
			cfg.CertFilePath,
			cfg.PrivateCryptoKey,
		)
		if err != nil {
			return nil, fmt.Errorf("an occured error when loading TLS keys: %s", err)
		}
		return creds, nil
	} else {
		creds := insecure.NewCredentials()
		return creds, nil
	}
}

func convertToBytes(val any) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(val)
	if err != nil {
		return nil, fmt.Errorf("an occured error when convert val to bytes, err: %w", err)
	}
	return buff.Bytes(), nil
}

func responseSize(val any) (int, error) {
	b, err := convertToBytes(val)
	if err != nil {
		return 0, fmt.Errorf("an occured error when calculate val size, err: %w", err)
	}
	return binary.Size(b), nil
}
