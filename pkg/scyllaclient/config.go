// Copyright (C) 2017 ScyllaDB

package scyllaclient

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// Config specifies the Client configuration.
type Config struct {
	// Hosts specifies all the cluster hosts that for a pool of hosts for the
	// client.
	Hosts []string
	// Port specifies the default Scylla Manager agent port.
	Port string
	// Transport scheme HTTP or HTTPS.
	Scheme string
	// AuthToken specifies the authentication token.
	AuthToken string
	// Timeout specifies end-to-end time to complete Scylla REST API request
	// including retries.
	Timeout time.Duration
	// RequestTimeout specifies time to complete a single request to Scylla
	// REST API possibly including opening a TCP connection.
	RequestTimeout time.Duration
	// PoolDecayDuration specifies size of time window to measure average
	// request time in Epsilon-Greedy host pool.
	PoolDecayDuration time.Duration

	// Transport allows for setting a custom round tripper to send HTTP requests
	// over not standard connections i.e. over SSH tunnel.
	Transport http.RoundTripper
}

// DefaultConfig returns a Config initialized with default values.
func DefaultConfig() Config {
	return Config{
		Port:              "10001",
		Scheme:            "https",
		Timeout:           90 * time.Second,
		RequestTimeout:    15 * time.Second,
		PoolDecayDuration: 30 * time.Minute,
	}
}

// TestConfig is a convenience function equal to calling DefaultConfig and
// setting hosts and token manually.
func TestConfig(hosts []string, token string) Config {
	config := DefaultConfig()
	config.Hosts = hosts
	config.AuthToken = token
	return config
}

// Validate checks if all the fields are properly set.
func (c Config) Validate() error {
	var err error
	if len(c.Hosts) == 0 {
		err = multierr.Append(err, errors.New("missing hosts"))
	}
	if c.Port == "" {
		err = multierr.Append(err, errors.New("missing port"))
	}

	return err
}
