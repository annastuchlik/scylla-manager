// Copyright (C) 2017 ScyllaDB

// +build all integration

package backup

import (
	"fmt"
	"testing"

	"github.com/scylladb/gocqlx"
	"github.com/scylladb/mermaid/pkg/schema/table"
	. "github.com/scylladb/mermaid/pkg/testutils"
	"github.com/scylladb/mermaid/pkg/util/uuid"
)

func TestRunProgressIteratorIntegration(t *testing.T) {
	expected := 100

	var (
		clusterID = uuid.NewTime()
		taskID    = uuid.NewTime()
		runID     = uuid.NewTime()
	)

	run := &Run{
		ID:        runID,
		ClusterID: clusterID,
		TaskID:    taskID,
	}

	p := &RunProgress{
		RunID:     runID,
		ClusterID: clusterID,
		TaskID:    taskID,
	}

	var prog []RunProgress
	session := CreateSession(t)

	stmt, names := table.BackupRunProgress.Insert()
	for i := 0; i < expected; i++ {
		p.Host = "host"
		p.Unit = int64(i)
		p.TableName = fmt.Sprintf("table_%d", i)
		q := gocqlx.Query(session.Query(stmt), names).BindStruct(p)
		if err := q.ExecRelease(); err != nil {
			t.Fatal(err)
		}
	}

	v := NewProgressVisitor(run, session)
	err := v.ForEach(func(pr *RunProgress) {
		prog = append(prog, *pr)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(prog) != expected {
		t.Fatalf("Expected %d items, got %d", expected, len(prog))
	}
	for i := 0; i < len(prog); i++ {
		if prog[i].Unit != int64(i) {
			t.Log(prog[i].TableName)
			t.Errorf("Expected Unit = %d, got %d", i, prog[i].Unit)
		}
	}
}
