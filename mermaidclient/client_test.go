// Copyright (C) 2017 ScyllaDB

package mermaidclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/google/go-cmp/cmp"
)

func TestClientError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "{\"message\": \"bla\"}", 500)
	}))
	defer s.Close()

	c, err := NewClient(s.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	_, err = c.ListClusters(ctx)
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*runtime.APIError)
	if !ok {
		t.Fatal("expected APIError")
	}

	if diff := cmp.Diff(string(apiErr.Response.(json.RawMessage)), "{\"message\": \"bla\"}"); diff != "" {
		t.Fatal(diff)
	}
}
