// Copyright (C) 2017 ScyllaDB

package healthcheck

const (
	statusUp   = `UP`
	statusDown = `DOWN`
)

// Status represents the status of a particular host
type Status struct {
	DC        string
	Host      string
	SSL       bool
	CQLStatus string  `json:"cql_status"`
	RTT       float64 `json:"cql_rtt_ms"`
}
