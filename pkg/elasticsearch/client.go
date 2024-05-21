package elasticsearch

import (
	"crypto/tls"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"net/http"
)

var Client *elasticsearch.TypedClient

func NewElasticsearchClient(cfg Config) (*elasticsearch.TypedClient, error) {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load x509 key pair: %v", err)
		}

		transport = http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
	}

	es, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		RetryOnStatus: []int{
			http.StatusTooManyRequests,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		Transport: &transport,
	})
	if err != nil {
		return nil, fmt.Errorf("init Elasticsearch client: %w", err)
	}

	Client = es
	return es, nil
}
