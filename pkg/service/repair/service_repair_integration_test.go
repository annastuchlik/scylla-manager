// Copyright (C) 2017 ScyllaDB

// +build all integration

package repair_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/scylladb/go-log"
	"github.com/scylladb/mermaid/pkg/scyllaclient"
	"github.com/scylladb/mermaid/pkg/service"
	"github.com/scylladb/mermaid/pkg/service/repair"
	. "github.com/scylladb/mermaid/pkg/testutils"
	"github.com/scylladb/mermaid/pkg/util/httpmw"
	"github.com/scylladb/mermaid/pkg/util/uuid"
	"go.uber.org/zap/zapcore"
)

// Cluster nodes
const (
	node0 = 0
	node1 = 1
	node2 = 2

	allShards = -1
)

type repairTestHelper struct {
	session *gocql.Session
	hrt     *HackableRoundTripper
	client  *scyllaclient.Client
	service *repair.Service

	clusterID uuid.UUID
	taskID    uuid.UUID
	runID     uuid.UUID

	done   bool
	result error
	mu     sync.Mutex

	t *testing.T
}

func newRepairTestHelper(t *testing.T, session *gocql.Session, config repair.Config) *repairTestHelper {
	t.Helper()

	ExecStmt(t, session, "TRUNCATE TABLE repair_run")
	ExecStmt(t, session, "TRUNCATE TABLE repair_run_progress")

	logger := log.NewDevelopmentWithLevel(zapcore.InfoLevel)

	hrt := NewHackableRoundTripper(scyllaclient.DefaultTransport())
	hrt.SetInterceptor(repairInterceptor(scyllaclient.CommandSuccessful))
	c := newTestClient(t, hrt, logger)
	s := newTestService(t, session, c, config, logger)

	return &repairTestHelper{
		session: session,
		hrt:     hrt,
		client:  c,
		service: s,

		clusterID: uuid.MustRandom(),
		taskID:    uuid.MustRandom(),
		runID:     uuid.NewTime(),

		t: t,
	}
}

func (h *repairTestHelper) runRepair(ctx context.Context, t repair.Target) {
	go func() {
		err := h.service.Repair(ctx, h.clusterID, h.taskID, h.runID, t)

		h.mu.Lock()
		h.done = true
		h.result = err
		h.mu.Unlock()
	}()
}

// WaitCond parameters
const (
	// now specifies that condition shall be true in the current state.
	now = 0
	// shortWait specifies that condition shall be met in immediate future
	// such as repair filing on start.
	shortWait = 4 * time.Second
	// longWait specifies that condition shall be met after a while, this is
	// useful for waiting for repair to significantly advance or finish.
	longWait = 20 * time.Second

	_interval = 100 * time.Millisecond
)

func (h *repairTestHelper) assertRunning(wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		_, err := h.service.GetProgress(context.Background(), h.clusterID, h.taskID, h.runID)
		if err != nil {
			if err == service.ErrNotFound {
				return false
			}
			h.t.Fatal(err)
		}
		return true
	}, _interval, wait)
}

func (h *repairTestHelper) assertDone(wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		h.mu.Lock()
		defer h.mu.Unlock()
		return h.done && h.result == nil
	}, _interval, wait)
}

func (h *repairTestHelper) assertProgressSuccess() {
	p, err := h.service.GetProgress(context.Background(), h.clusterID, h.taskID, h.runID)
	if err != nil {
		h.t.Fatal(err)
	}

	se := 0
	ss := 0
	st := 0
	for _, u := range p.Units {
		for _, n := range u.Nodes {
			for _, s := range n.Shards {
				se += s.SegmentError
				ss += s.SegmentSuccess
				st += s.SegmentCount
			}
		}
	}

	Print("And: there are no more errors")
	if se != 0 {
		h.t.Fatal("expected", 0, "got", se)
	}
	if ss != st {
		h.t.Fatal("expected", st, "got", ss)
	}
}

func (h *repairTestHelper) assertError(wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		h.mu.Lock()
		defer h.mu.Unlock()
		return h.done && h.result != nil
	}, _interval, wait)
}

func (h *repairTestHelper) assertErrorContains(cause string, wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		h.mu.Lock()
		defer h.mu.Unlock()
		return h.done && h.result != nil && strings.Contains(h.result.Error(), cause)
	}, _interval, wait)
}

func (h *repairTestHelper) assertStopped(wait time.Duration) {
	h.t.Helper()
	h.assertErrorContains(context.Canceled.Error(), wait)
}

func (h *repairTestHelper) assertProgress(unit, node, percent int, wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		p, _ := h.progress(unit, node, allShards)
		return p >= percent
	}, _interval, wait)
}

func (h *repairTestHelper) assertProgressFailed(unit, node, percent int, wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		_, f := h.progress(unit, node, allShards)
		return f >= percent
	}, _interval, wait)
}

func (h *repairTestHelper) assertMaxProgress(unit, node, percent int, wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		p, _ := h.progress(unit, node, allShards)
		return p <= percent
	}, _interval, wait)
}

func (h *repairTestHelper) assertShardProgress(unit, node, shard, percent int, wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		p, _ := h.progress(unit, node, shard)
		return p >= percent
	}, _interval, wait)
}

func (h *repairTestHelper) assertMaxShardProgress(unit, node, shard, percent int, wait time.Duration) {
	h.t.Helper()

	WaitCond(h.t, func() bool {
		p, _ := h.progress(unit, node, shard)
		return p <= percent
	}, _interval, wait)
}

func (h *repairTestHelper) progress(unit, node, shard int) (int, int) {
	h.t.Helper()

	p, err := h.service.GetProgress(context.Background(), h.clusterID, h.taskID, h.runID)
	if err != nil {
		h.t.Fatal(err)
	}

	if len(p.Units[unit].Nodes) <= node {
		return -1, -1
	}

	if shard < 0 {
		return int(p.Units[unit].Nodes[node].PercentComplete), int(p.Units[unit].Nodes[node].PercentFailed)
	}

	if len(p.Units[unit].Nodes[node].Shards) <= shard {
		return -1, -1
	}

	return int(p.Units[unit].Nodes[node].Shards[shard].PercentComplete), int(p.Units[unit].Nodes[node].Shards[shard].PercentFailed)
}

func newTestClient(t *testing.T, hrt *HackableRoundTripper, logger log.Logger) *scyllaclient.Client {
	t.Helper()

	config := scyllaclient.TestConfig(ManagedClusterHosts(), AgentAuthToken())
	config.Transport = hrt
	c, err := scyllaclient.NewClient(config, logger.Named("scylla"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newTestService(t *testing.T, session *gocql.Session, client *scyllaclient.Client, c repair.Config, logger log.Logger) *repair.Service {
	t.Helper()

	s, err := repair.NewService(
		session,
		c,
		func(_ context.Context, id uuid.UUID) (string, error) {
			return "test_cluster", nil
		},
		func(context.Context, uuid.UUID) (*scyllaclient.Client, error) {
			return client, nil
		},
		logger.Named("repair"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

var commandCounter int32

func repairInterceptor(s scyllaclient.CommandStatus) http.RoundTripper {
	return httpmw.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(req.URL.Path, "/storage_service/repair_async/") {
			return nil, nil
		}

		resp := &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Request:    req,
			Header:     make(http.Header, 0),
		}

		switch req.Method {
		case http.MethodGet:
			resp.Body = ioutil.NopCloser(bytes.NewBufferString(fmt.Sprintf("\"%s\"", s)))
		case http.MethodPost:
			id := atomic.AddInt32(&commandCounter, 1)
			resp.Body = ioutil.NopCloser(bytes.NewBufferString(fmt.Sprint(id)))
		}

		return resp, nil
	})
}

func unstableRepairInterceptor() http.RoundTripper {
	failRi := repairInterceptor(scyllaclient.CommandFailed)
	successRi := repairInterceptor(scyllaclient.CommandSuccessful)
	return httpmw.RoundTripperFunc(func(req *http.Request) (resp *http.Response, err error) {
		id := atomic.LoadInt32(&commandCounter)
		if id != 0 && id%20 == 0 {
			return failRi.RoundTrip(req)
		}
		return successRi.RoundTrip(req)
	})
}

func dialErrorInterceptor() http.RoundTripper {
	return httpmw.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("mock dial error")
	})
}

func createKeyspace(t *testing.T, session *gocql.Session, keyspace string) {
	ExecStmt(t, session, "CREATE KEYSPACE "+keyspace+" WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 3, 'dc2': 3}")
}

func dropKeyspace(t *testing.T, session *gocql.Session, keyspace string) {
	ExecStmt(t, session, "DROP KEYSPACE "+keyspace)
}

func singleUnit() repair.Target {
	return repair.Target{
		Units: []repair.Unit{
			{
				Keyspace: "test_repair",
				Tables:   []string{"test_table_0"},
			},
		},
		DC:          []string{"dc1", "dc2"},
		TokenRanges: repair.DCPrimaryTokenRanges,
		Continue:    true,
	}
}

func multipleUnits() repair.Target {
	return repair.Target{
		Units: []repair.Unit{
			{Keyspace: "test_repair", Tables: []string{"test_table_0"}},
			{Keyspace: "test_repair", Tables: []string{"test_table_1"}},
		},
		DC:          []string{"dc1", "dc2"},
		TokenRanges: repair.DCPrimaryTokenRanges,
		Continue:    true,
	}
}

func TestServiceRepairIntegration(t *testing.T) {
	clusterSession := CreateManagedClusterSession(t)

	createKeyspace(t, clusterSession, "test_repair")
	ExecStmt(t, clusterSession, "CREATE TABLE test_repair.test_table_0 (id int PRIMARY KEY)")
	ExecStmt(t, clusterSession, "CREATE TABLE test_repair.test_table_1 (id int PRIMARY KEY)")
	defer dropKeyspace(t, clusterSession, "test_repair")

	defaultConfig := func() repair.Config {
		c := repair.DefaultConfig()
		c.SegmentsPerRepair = 10
		c.PollInterval = 10 * time.Millisecond
		return c
	}

	session := CreateSession(t)

	t.Run("repair simple", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, multipleUnits())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node0 advances")
		h.assertProgress(0, node0, 1, shortWait)

		Print("When: node0 is 100% repaired")
		h.assertProgress(0, node0, 100, longWait)

		Print("Then: repair of node1 advances")
		h.assertProgress(0, node1, 1, shortWait)

		Print("When: node1 is 100% repaired")
		h.assertProgress(0, node1, 100, longWait)

		Print("Then: repair of node2 advances")
		h.assertProgress(0, node2, 1, shortWait)

		Print("When: node2 is 100% repaired")
		h.assertProgress(0, node2, 100, longWait)

		Print("And: repair of U1 node0 advances")
		h.assertProgress(1, node0, 1, shortWait)

		Print("When: U1 node0 is 100% repaired")
		h.assertProgress(1, node0, 100, longWait)

		Print("Then: repair of U1 node1 advances")
		h.assertProgress(1, node1, 1, shortWait)

		Print("When: U1 node1 is 100% repaired")
		h.assertProgress(1, node1, 100, longWait)

		Print("Then: repair of U1 node2 advances")
		h.assertProgress(1, node2, 1, shortWait)

		Print("When: U1 node2 is 100% repaired")
		h.assertProgress(1, node2, 100, longWait)

		Print("Then: repair is done")
		h.assertDone(shortWait)
	})

	t.Run("repair dc", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		units := multipleUnits()
		units.DC = []string{"dc2"}

		Print("When: run repair")
		h.runRepair(ctx, units)

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: repair is done")
		h.assertDone(2 * longWait)

		Print("Then: dc2 is used for repair")
		prog, err := h.service.GetProgress(context.Background(), h.clusterID, h.taskID, h.runID)
		if err != nil {
			h.t.Fatal(err)
		}
		if diff := cmp.Diff(prog.DC, []string{"dc2"}); diff != "" {
			h.t.Fatal(diff)
		}
		for _, u := range prog.Units {
			for _, n := range u.Nodes {
				if !strings.HasPrefix(n.Host, "192.168.100.2") {
					t.Error(n.Host)
				}
			}
		}
	})

	t.Run("repair dc local keyspace mismatch", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("Given: dc2 only keyspace")
		ExecStmt(t, clusterSession, "CREATE KEYSPACE IF NOT EXISTS test_repair_dc2 WITH replication = {'class': 'NetworkTopologyStrategy', 'dc2': 3}")
		ExecStmt(t, clusterSession, "CREATE TABLE IF NOT EXISTS test_repair_dc2.test_table_0 (id int PRIMARY KEY)")

		units := singleUnit()
		units.Units = []repair.Unit{
			{
				Keyspace: "test_repair_dc2",
			},
		}
		units.DC = []string{"dc1"}

		Print("When: run repair with dc1")
		h.runRepair(ctx, units)

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: repair fails")
		h.assertErrorContains("no matching DCs", shortWait)
	})

	t.Run("repair simple strategy multi dc", func(t *testing.T) {
		systemAuthUnit := repair.Target{
			Units: []repair.Unit{
				{
					Keyspace: "system_auth",
				},
			},
			DC:          []string{"dc1"},
			TokenRanges: repair.DCPrimaryTokenRanges,
			Continue:    true,
		}

		const (
			node3 = 3
			node4 = 4
			node5 = 5
		)

		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, systemAuthUnit)

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node0 advances")
		h.assertProgress(0, node0, 1, shortWait)

		Print("When: node0 is 100% repaired")
		h.assertProgress(0, node0, 100, longWait)

		Print("Then: repair of node1 advances")
		h.assertProgress(0, node1, 1, shortWait)

		Print("When: node1 is 100% repaired")
		h.assertProgress(0, node1, 100, longWait)

		Print("Then: repair of node2 advances")
		h.assertProgress(0, node2, 1, shortWait)

		Print("When: node2 is 100% repaired")
		h.assertProgress(0, node2, 100, longWait)

		Print("When: node3 is 100% repaired")
		h.assertProgress(0, node3, 100, longWait)

		Print("When: node4 is 100% repaired")
		h.assertProgress(0, node4, 100, longWait)

		Print("When: node5 is 100% repaired")
		h.assertProgress(0, node5, 100, longWait)

		Print("Then: repair is done")
		h.assertDone(shortWait)
	})

	t.Run("repair host", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		units := multipleUnits()
		units.Host = ManagedClusterHost()

		Print("When: run repair")
		h.runRepair(ctx, units)

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node0 advances")
		h.assertProgress(0, node0, 1, shortWait)

		Print("When: node0 is 100% repaired")
		h.assertProgress(0, node0, 100, longWait)

		Print("Then: repair is done")
		h.assertDone(2 * longWait)

		Print(fmt.Sprintf("Then: host %s is used for repair", units.Host))
		prog, err := h.service.GetProgress(context.Background(), h.clusterID, h.taskID, h.runID)
		if err != nil {
			h.t.Fatal(err)
		}
		for _, u := range prog.Units {
			for _, n := range u.Nodes {
				if n.Host != units.Host {
					t.Error(n.Host)
				}
			}
		}
	})

	t.Run("repair shard limit", func(t *testing.T) {
		c := defaultConfig()
		c.ShardParallelMax = 1

		h := newRepairTestHelper(t, session, c)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, singleUnit())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node0 shard 0 advances")
		h.assertShardProgress(0, node0, 0, 1, shortWait)

		Print("When: node0 shard 0 is 50% repaired")
		h.assertShardProgress(0, node0, 0, 50, longWait)

		Print("Then: node0 shard 1 did not start")
		h.assertMaxShardProgress(0, node0, 1, 0, now)

		Print("When: node0 shard 0 is 100% repaired")
		h.assertShardProgress(0, node0, 0, 100, longWait)

		Print("Then: repair of node0 shard 1 advances")
		h.assertShardProgress(0, node0, 1, 1, shortWait)

		Print("When: node0 shard 1 is 100% repaired")
		h.assertShardProgress(0, node0, 1, 100, longWait)

		Print("Then: repair of node1 shard 0 advances")
		h.assertShardProgress(0, node1, 0, 1, shortWait)

		Print("When: node1 shard 0 is 50% repaired")
		h.assertShardProgress(0, node1, 0, 50, longWait)

		Print("Then: node1 shard 1 did not start")
		h.assertMaxShardProgress(0, node1, 1, 0, now)

		Print("When: node1 shard 0 is 100% repaired")
		h.assertShardProgress(0, node1, 0, 100, longWait)

		Print("Then: repair of node1 shard 1 advances")
		h.assertShardProgress(0, node1, 1, 1, shortWait)

		Print("When: node1 shard 1 is 100% repaired")
		h.assertShardProgress(0, node1, 1, 100, longWait)

		Print("Then: repair of node2 shard 0 advances")
		h.assertShardProgress(0, node2, 0, 1, shortWait)

		Print("When: node2 shard 0 is 50% repaired")
		h.assertShardProgress(0, node2, 0, 50, longWait)

		Print("Then: node2 shard 1 did not start")
		h.assertMaxShardProgress(0, node2, 1, 0, now)

		Print("When: node2 shard 0 is 100% repaired")
		h.assertShardProgress(0, node2, 0, 100, longWait)

		Print("Then: repair of node2 shard 1 advances")
		h.assertShardProgress(0, node2, 1, 1, shortWait)

		Print("When: node2 shard 1 is 100% repaired")
		h.assertShardProgress(0, node2, 1, 100, longWait)

		Print("Then: repair is done")
		h.assertDone(shortWait)
	})

	t.Run("repair stop", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, singleUnit())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: node0 is 50% repaired")
		h.assertProgress(0, node0, 50, longWait)

		Print("And: stop repair")
		cancel()

		Print("Then: status is StatusStopped")
		h.assertStopped(shortWait)
	})

	t.Run("repair stop while backoff", func(t *testing.T) {
		c := defaultConfig()
		c.ErrorBackoff = 1000 * longWait
		h := newRepairTestHelper(t, session, c)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, singleUnit())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: repair of node0 advances")
		h.assertProgress(0, node0, 1, shortWait)

		Print("And: error occurs")
		h.hrt.SetInterceptor(repairInterceptor(scyllaclient.CommandFailed))

		Print("And: backoff kicks in")
		time.Sleep(shortWait)

		Print("And: stop repair")
		cancel()

		Print("And: status is StatusStopped")
		h.assertStopped(shortWait)
	})

	t.Run("repair restart", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, multipleUnits())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: node1 is 50% repaired")
		h.assertProgress(0, node1, 50, longWait)

		Print("And: stop repair")
		cancel()

		Print("Then: status is StatusStopped")
		h.assertStopped(shortWait)

		Print("When: create a new task")
		h.runID = uuid.NewTime()

		Print("And: run repair")
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		h.runRepair(ctx, multipleUnits())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node1 continues")
		h.assertProgress(0, node0, 100, shortWait)
		h.assertProgress(0, node1, 50, now)
		h.assertProgress(0, node2, 0, now)

		Print("When: U1 node0 is 10% repaired")
		h.assertProgress(1, node0, 10, longWait)

		Print("And: stop repair")
		cancel()

		Print("Then: status is StatusStopped")
		h.assertStopped(shortWait)

		Print("When: create a new task")
		h.runID = uuid.NewTime()

		Print("And: run repair")
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		h.runRepair(ctx, multipleUnits())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of U1 node0 continues")
		h.assertProgress(0, node0, 100, shortWait)
		h.assertProgress(0, node1, 100, now)
		h.assertProgress(0, node2, 100, now)
		h.assertProgress(1, node0, 10, shortWait)
	})

	t.Run("repair restart no continue", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		unit := singleUnit()
		unit.Continue = false

		Print("When: run repair")
		h.runRepair(ctx, unit)

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: node0 is 50% repaired")
		h.assertProgress(0, node0, 50, longWait)

		Print("And: stop repair")
		cancel()

		Print("Then: status is StatusStopped")
		h.assertStopped(shortWait)

		Print("When: create a new task")
		h.runID = uuid.NewTime()

		Print("And: run repair")
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		h.runRepair(ctx, unit)

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node0 starts from scratch")
		h.assertProgress(0, node0, 1, shortWait)
		if p, _ := h.progress(0, node0, allShards); p >= 50 {
			t.Fatal("node0 should start from schratch")
		}
	})

	t.Run("repair restart task properties changed", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, multipleUnits())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: node1 is 50% repaired")
		h.assertProgress(0, node1, 50, longWait)

		Print("And: stop repair")
		cancel()

		Print("Then: status is StatusStopped")
		h.assertStopped(shortWait)

		Print("When: create a new task")
		h.runID = uuid.NewTime()

		Print("And: run repair with modified units")
		modifiedUnits := multipleUnits()
		modifiedUnits.Units = []repair.Unit{{Keyspace: "test_repair", Tables: []string{"test_table_1"}}}
		modifiedUnits.DC = []string{"dc2"}

		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		h.runRepair(ctx, modifiedUnits)

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node1 continues")
		h.assertProgress(0, node0, 100, shortWait)
		h.assertProgress(0, node1, 50, now)
		h.assertProgress(0, node2, 0, now)

		Print("And: dc1 is used for repair")
		prog, err := h.service.GetProgress(context.Background(), h.clusterID, h.taskID, h.runID)
		if err != nil {
			h.t.Fatal(err)
		}
		if diff := cmp.Diff(prog.DC, multipleUnits().DC); diff != "" {
			h.t.Fatal(diff, prog)
		}
		if len(prog.Units) != 2 {
			t.Fatal(prog.Units)
		}
	})

	t.Run("repair restart task segments_per_repair changed", func(t *testing.T) {
		c := defaultConfig()
		c.ErrorBackoff = time.Millisecond
		c.SegmentsPerRepair = 3
		h := newRepairTestHelper(t, session, c)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h.hrt.SetInterceptor(unstableRepairInterceptor())

		Print("When: run repair")
		h.runRepair(ctx, multipleUnits())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: node0 is 5% repaired")
		h.assertProgress(0, node0, 5, shortWait)

		Print("And: node0 is 2% failed")
		h.assertProgressFailed(0, node0, 2, shortWait)

		Print("And: stop repair")
		cancel()

		Print("Then: status is StatusStopped")
		h.assertStopped(shortWait)

		Print("When: create a new task")
		c.SegmentsPerRepair = 7
		h = newRepairTestHelper(t, session, c)
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()

		Print("And: run repair with modified segments_per_repair")
		h.runRepair(ctx, multipleUnits())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("And: repair of node1 continues")
		h.assertProgress(0, node0, 100, shortWait)
		h.assertProgress(0, node1, 50, shortWait)
		h.assertProgress(0, node2, 50, shortWait)

		Print("Then: status is StatusDone")
		h.assertDone(longWait)

		Print("And: dc1 is used for repair")
		prog, err := h.service.GetProgress(context.Background(), h.clusterID, h.taskID, h.runID)
		if err != nil {
			h.t.Fatal(err)
		}
		if diff := cmp.Diff(prog.DC, multipleUnits().DC); diff != "" {
			h.t.Fatal(diff, prog)
		}
		if len(prog.Units) != 2 {
			t.Fatal(prog.Units)
		}

		Print("And: progress is successful")
		h.assertProgressSuccess()
	})

	t.Run("repair temporary network outage", func(t *testing.T) {
		c := defaultConfig()
		c.ErrorBackoff = 1 * time.Second

		h := newRepairTestHelper(t, session, c)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, singleUnit())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: node1 is 50% repaired")
		h.assertProgress(0, node1, 50, longWait)

		Print("And: no network for 5s with 1s backoff")
		h.hrt.SetInterceptor(dialErrorInterceptor())
		time.AfterFunc(3*h.client.Config().Timeout, func() {
			h.hrt.SetInterceptor(repairInterceptor(scyllaclient.CommandSuccessful))
		})

		Print("Then: node1 repair continues")
		h.assertProgress(0, node1, 60, 3*h.client.Config().Timeout+longWait)

		Print("When: node1 is 95% repaired")
		h.assertProgress(0, node1, 95, longWait)

		Print("Then: repair of node2 advances")
		h.assertProgress(0, node2, 1, longWait)

		Print("When: node2 is 100% repaired")
		h.assertProgress(0, node2, 100, longWait)

		Print("Then: node1 is retries repair")
		h.assertProgress(0, node1, 100, shortWait)

		Print("And: repair is done")
		h.assertDone(shortWait)
	})

	t.Run("repair error retry", func(t *testing.T) {
		c := defaultConfig()
		c.ErrorBackoff = 1 * time.Second

		h := newRepairTestHelper(t, session, c)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("When: run repair")
		h.runRepair(ctx, singleUnit())

		Print("Then: repair is running")
		h.assertRunning(shortWait)

		Print("When: node1 is 50% repaired")
		h.assertProgress(0, node1, 50, longWait)

		Print("And: errors occur for 5s with 1s backoff")
		h.hrt.SetInterceptor(repairInterceptor(scyllaclient.CommandFailed))
		time.AfterFunc(5*time.Second, func() {
			h.hrt.SetInterceptor(repairInterceptor(scyllaclient.CommandSuccessful))
		})

		Print("Then: node1 repair continues")
		h.assertProgress(0, node1, 60, longWait)

		Print("When: node1 is 95% repaired")
		h.assertProgress(0, node1, 95, longWait)

		Print("Then: repair of node2 advances")
		h.assertProgress(0, node2, 1, longWait)

		Print("When: node2 is 100% repaired")
		h.assertProgress(0, node2, 100, longWait)

		Print("Then: node1 is retries repair")
		h.assertProgress(0, node1, 100, shortWait)

		Print("And: repair is done")
		h.assertDone(shortWait)
	})

	t.Run("repair error fail fast", func(t *testing.T) {
		c := defaultConfig()
		h := newRepairTestHelper(t, session, c)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		unit := singleUnit()
		unit.FailFast = true

		Print("Given: one repair fails")
		var (
			mu sync.Mutex
			ic = repairInterceptor(scyllaclient.CommandFailed)
		)
		h.hrt.SetInterceptor(httpmw.RoundTripperFunc(func(req *http.Request) (resp *http.Response, err error) {
			if !strings.HasPrefix(req.URL.Path, "/storage_service/repair_async/") {
				return nil, nil
			}

			mu.Lock()
			defer mu.Unlock()
			resp, err = ic.RoundTrip(req)

			if req.Method == http.MethodGet {
				ic = repairInterceptor(scyllaclient.CommandSuccessful)
			}
			return
		}))

		Print("And: repair")
		h.runRepair(ctx, unit)

		Print("Then: repair fails")
		h.assertError(shortWait)

		Print("And: errors are recorded")
		p, err := h.service.GetProgress(ctx, h.clusterID, h.taskID, h.runID)
		if err != nil {
			t.Fatal(err)
		}

		se := 0
		ss := 0
		for _, u := range p.Units {
			for _, n := range u.Nodes {
				for _, s := range n.Shards {
					se += s.SegmentError
					ss += s.SegmentSuccess
				}
			}
		}
		if se != c.SegmentsPerRepair {
			t.Fatal("expected", c.SegmentsPerRepair, "got", se)
		}
		if ss > 10*c.SegmentsPerRepair {
			t.Fatal("got", ss) // sometimes can be 0 or 20
		}
	})

	t.Run("repair non existing keyspace", func(t *testing.T) {
		h := newRepairTestHelper(t, session, defaultConfig())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		Print("Given: non-existing keyspace")

		target := singleUnit()
		target.Units[0].Keyspace = "non_existing_keyspace"

		Print("When: run repair")
		h.runRepair(ctx, target)

		Print("Then: repair fails")
		h.assertError(shortWait)
	})
}

func TestServiceRepairErrorNodetoolRepairRunningIntegration(t *testing.T) {
	clusterSession := CreateManagedClusterSession(t)

	createKeyspace(t, clusterSession, "test_repair")
	ExecStmt(t, clusterSession, "CREATE TABLE test_repair.test_table_0 (id int PRIMARY KEY)")
	ExecStmt(t, clusterSession, "CREATE TABLE test_repair.test_table_1 (id int PRIMARY KEY)")
	defer dropKeyspace(t, clusterSession, "test_repair")

	session := CreateSession(t)
	h := newRepairTestHelper(t, session, repair.DefaultConfig())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Print("Given: repair is running on a host")
	done := make(chan struct{})
	go func() {
		time.AfterFunc(500*time.Millisecond, func() {
			close(done)
		})
		ExecOnHost(ManagedClusterHost(), "nodetool repair -pr")
	}()
	defer func() {
		if err := h.client.KillAllRepairs(context.Background(), ManagedClusterHost()); err != nil {
			t.Fatal(err)
		}
	}()

	<-done
	Print("When: repair starts")
	h.runRepair(ctx, singleUnit())

	Print("Then: repair fails")
	h.assertErrorContains("active repair on hosts", longWait)
}

func TestServiceGetTargetSkipsKeyspaceHavingNoReplicasInGivenDCIntegration(t *testing.T) {
	clusterSession := CreateManagedClusterSession(t)

	ExecStmt(t, clusterSession, "CREATE KEYSPACE test_repair_0 WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 3, 'dc2': 3}")
	ExecStmt(t, clusterSession, "CREATE TABLE test_repair_0.test_table_0 (id int PRIMARY KEY)")
	defer dropKeyspace(t, clusterSession, "test_repair_0")

	ExecStmt(t, clusterSession, "CREATE KEYSPACE test_repair_1 WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 3, 'dc2': 0}")
	ExecStmt(t, clusterSession, "CREATE TABLE test_repair_1.test_table_0 (id int PRIMARY KEY)")
	defer dropKeyspace(t, clusterSession, "test_repair_1")

	session := CreateSession(t)
	h := newRepairTestHelper(t, session, repair.DefaultConfig())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	properties := map[string]interface{}{
		"keyspace": []string{"test_repair_0", "test_repair_1"},
		"dc":       []string{"dc2"},
	}

	props, err := json.Marshal(properties)
	if err != nil {
		t.Fatal(err)
	}

	target, err := h.service.GetTarget(ctx, h.clusterID, props, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(target.Units) != 1 {
		t.Errorf("Expected single unit in target, get %d", len(target.Units))
	}
	if target.Units[0].Keyspace != "test_repair_0" {
		t.Errorf("Expected only 'test_repair_0' keyspace in target units, got %s", target.Units[0].Keyspace)
	}
}
