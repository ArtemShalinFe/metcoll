// Package metcoll for interacting with the metrics server.
package metcoll

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

// Client - sends requests for metric updates to the server.
type Client struct {
	host       string
	httpClient *retryablehttp.Client
	logger     retryablehttp.LeveledLogger
	hashkey    []byte
}

const (
	HashSHA256 = "HashSHA256"
)

// NewClient - Object constructor.
func NewClient(cfg *configuration.ConfigAgent, logger retryablehttp.LeveledLogger) *Client {
	const defautMaxRetry = 3
	const defautMinWaitRetry = 3 * time.Second
	const defautMaxWaitRetry = 5 * time.Second

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = defautMaxRetry
	retryClient.CheckRetry = checkRetry
	retryClient.RetryWaitMin = time.Duration(defautMinWaitRetry)
	retryClient.RetryWaitMax = time.Duration(defautMaxWaitRetry)
	retryClient.Logger = logger
	retryClient.Backoff = backoff

	return &Client{
		host:       cfg.Server,
		httpClient: retryClient,
		logger:     logger,
		hashkey:    cfg.Key,
	}
}

func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true, err
	} else {
		return false, err
	}
}

func backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	const an0 = 0
	const an1 = 1
	const an2 = 2

	const an0backoff = 1 * time.Second
	const an1backoff = 3 * time.Second
	const an2backoff = 5 * time.Second
	const defaultbackoff = 2 * time.Second

	switch attemptNum {
	case an0:
		return an0backoff
	case an1:
		return an1backoff
	case an2:
		return an2backoff
	default:
		return defaultbackoff
	}
}

func (c *Client) prepareRequest(ctx context.Context, body []byte, url string) (*retryablehttp.Request, error) {
	var zBuf bytes.Buffer
	zw := gzip.NewWriter(&zBuf)

	if _, err := zw.Write(body); err != nil {
		return nil, fmt.Errorf("cannot write compressed body err: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("cannot close compress writer err: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, url, &zBuf)
	if err != nil {
		return nil, fmt.Errorf("cannot create request err: %w", err)
	}
	req.Close = true

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if len(c.hashkey) != 0 {
		data, err := req.BodyBytes()
		if err != nil {
			return nil, fmt.Errorf("cannot calculate hash err: %w", err)
		}

		h := hmac.New(sha256.New, []byte(c.hashkey))

		h.Write(data)

		req.Header.Set(HashSHA256, fmt.Sprintf("%x", h.Sum(nil)))
	}

	return req, nil
}

func (c *Client) batchUpdate(ctx context.Context, metrics []*metrics.Metrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("cannot marshal metric err: %w", err)
	}

	url, err := url.JoinPath("http://", c.host, "/updates/")
	if err != nil {
		return fmt.Errorf("cannot join elements in path err: %w", err)
	}

	req, err := c.prepareRequest(ctx, body, url)
	if err != nil {
		return fmt.Errorf("cannot prepare request err: %w", err)
	}

	return c.doRequest(req)
}

func (c *Client) doRequest(req *retryablehttp.Request) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request execute err: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("an error occured while body closing err: %v", err)
		}
	}()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body err: %w", err)
	}

	if resp.StatusCode < http.StatusMultipleChoices {
		c.logger.Info(`request for update metric has been completed
	code: %d, hash: %s`, resp.StatusCode, resp.Header.Get(HashSHA256))
	} else {
		c.logger.Error(`request for update metric has failed
	code: %d
	result: %s`, resp.StatusCode, string(res))
	}

	return nil
}

// PushResult - Consolidates the result of sending the metric and the error.
type PushResult struct {
	Metric *metrics.Metrics
	Err    error
}

// BatchUpdateMetric - Sends updated metrics received from the channel `mcs` to the server.
func (c *Client) BatchUpdateMetric(ctx context.Context, mcs <-chan []*metrics.Metrics, result chan<- error) {
	for m := range mcs {
		if err := c.batchUpdate(ctx, m); err != nil {
			result <- err
		}
	}

	select {
	case <-ctx.Done():
		return
	default:
	}
}

func certs() {
	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
}
