// Copyright (C) 2017 ScyllaDB

package healthcheck

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/scylladb/go-log"
	"github.com/scylladb/mermaid"
	"github.com/scylladb/mermaid/cluster"
	"github.com/scylladb/mermaid/internal/cqlping"
	"github.com/scylladb/mermaid/internal/kv"
	"github.com/scylladb/mermaid/sched/runner"
	"github.com/scylladb/mermaid/scyllaclient"
	"github.com/scylladb/mermaid/uuid"
)

// Health check defaults.
var (
	DefaultPort        = 9042
	DefaultPingTimeout = 250 * time.Millisecond
	DefaultTLSConfig   = &tls.Config{
		InsecureSkipVerify: true,
	}
)

// Service manages health checks.
type Service struct {
	cluster      cluster.ProviderFunc
	client       scyllaclient.ProviderFunc
	sslCertStore kv.Store
	sslKeyStore  kv.Store
	cache        map[uuid.UUID]*tls.Config
	cacheMu      sync.Mutex
	logger       log.Logger
}

func NewService(cp cluster.ProviderFunc, sp scyllaclient.ProviderFunc,
	sslCertStore, sslKeyStore kv.Store, logger log.Logger) (*Service, error) {
	if cp == nil {
		return nil, errors.New("invalid cluster provider")
	}
	if sp == nil {
		return nil, errors.New("invalid scylla provider")
	}
	if sslCertStore == nil {
		return nil, errors.New("missing SSL cert store")
	}
	if sslKeyStore == nil {
		return nil, errors.New("missing SSL key store")
	}

	return &Service{
		cluster:      cp,
		client:       sp,
		sslCertStore: sslCertStore,
		sslKeyStore:  sslKeyStore,
		cache:        make(map[uuid.UUID]*tls.Config),
		logger:       logger,
	}, nil
}

// Runner creates a runner.Runner that performs health checks.
func (s *Service) Runner() runner.Runner {
	return healthCheckRunner{
		client:  s.client,
		cluster: s.cluster,
		ping:    s.ping,
	}
}

// GetStatus returns the current status of the supplied cluster.
func (s *Service) GetStatus(ctx context.Context, clusterID uuid.UUID) ([]Status, error) {
	s.logger.Debug(ctx, "GetStatus", "cluster_id", clusterID)

	client, err := s.client(ctx, clusterID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get client for cluster with id %s", clusterID)
	}

	dcs, err := client.Datacenters(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dcs for cluster with id %s", clusterID)
	}

	out := make(chan Status, runtime.NumCPU()+1)
	size := 0
	for dc, hosts := range dcs {
		for _, h := range hosts {
			v := Status{
				DC:   dc,
				Host: h,
			}
			size++

			go func() {
				rtt, err := s.ping(ctx, clusterID, v.Host)
				if err != nil {
					s.logger.Info(ctx, "Ping failed",
						"cluster_id", clusterID,
						"host", v.Host,
						"error", err,
					)
					v.CQLStatus = statusDown
				} else {
					v.CQLStatus = statusUp
				}
				v.SSL = s.hasTLSConfig(clusterID)
				v.RTT = float64(rtt / 1000000)

				out <- v
			}()
		}
	}

	statuses := make([]Status, size)
	for i := 0; i < size; i++ {
		statuses[i] = <-out
	}
	sort.Slice(statuses, func(i, j int) bool {
		if statuses[i].DC == statuses[j].DC {
			return statuses[i].Host < statuses[j].Host
		}
		return statuses[i].DC < statuses[j].DC
	})

	return statuses, nil
}

func (s *Service) ping(ctx context.Context, clusterID uuid.UUID, host string) (rtt time.Duration, err error) {
	tlsConfig, err := s.tlsConfig(ctx, clusterID)
	if err != nil {
		return 0, err
	}

	config := cqlping.Config{
		Addr:      fmt.Sprint(host, ":", DefaultPort),
		Timeout:   DefaultPingTimeout,
		TLSConfig: tlsConfig,
	}
	rtt, err = cqlping.Ping(ctx, config)

	// If connection was cut by the server try upgrading to TLS.
	if errors.Cause(err) == io.EOF && config.TLSConfig == nil {
		s.logger.Info(ctx, "Upgrading connection to TLS",
			"cluster_id", clusterID,
			"host", host,
		)
		config.TLSConfig = DefaultTLSConfig
		rtt, err = cqlping.Ping(ctx, config)
		if err == nil {
			s.setTLSConfig(clusterID, DefaultTLSConfig)
		}
	}

	return
}

func (s *Service) tlsConfig(ctx context.Context, clusterID uuid.UUID) (*tls.Config, error) {
	// Try loading from cache.
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	if c, ok := s.cache[clusterID]; ok {
		return c, nil
	}

	s.logger.Info(ctx, "Loading SSL certificate from secure store", "cluster_id", clusterID)
	cert, err := s.sslCertStore.Get(clusterID)
	// If there is no user certificate record no TLS config to avoid rereading
	// from a secure store.
	if err == mermaid.ErrNotFound {
		s.cache[clusterID] = nil
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get SSL user cert from a secure store")
	}
	key, err := s.sslKeyStore.Get(clusterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get SSL user key from a secure store")
	}
	keyPair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, errors.Wrap(err, "invalid SSL user key pair")
	}

	// Create a new TLS configuration with user certificate and cache
	// the configuration.
	cfg := &tls.Config{
		Certificates:       []tls.Certificate{keyPair},
		InsecureSkipVerify: true,
	}
	s.cache[clusterID] = cfg

	return cfg, nil
}

func (s *Service) setTLSConfig(clusterID uuid.UUID, config *tls.Config) {
	s.cacheMu.Lock()
	s.cache[clusterID] = config
	s.cacheMu.Unlock()
}

func (s *Service) hasTLSConfig(clusterID uuid.UUID) bool {
	s.cacheMu.Lock()
	_, ok := s.cache[clusterID]
	s.cacheMu.Unlock()
	return ok
}

// InvalidateTLSConfigCache frees all in-memory TLS configuration associated
// with a given cluster forcing reload from a key store with next usage.
func (s *Service) InvalidateTLSConfigCache(clusterID uuid.UUID) {
	s.cacheMu.Lock()
	delete(s.cache, clusterID)
	s.cacheMu.Unlock()
}
