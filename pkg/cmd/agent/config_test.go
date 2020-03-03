// Copyright (C) 2017 ScyllaDB

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

var update = flag.Bool("update", false, "update .golden files")

func TestParsingConfig(t *testing.T) {
	table := []struct {
		Name   string
		Input  []string
		Golden string
	}{
		{
			Name:   "basic",
			Input:  []string{"./testdata/config/basic.input.yaml"},
			Golden: "./testdata/config/basic.golden.yaml",
		},
		{
			Name:   "scylla overwrite",
			Input:  []string{"./testdata/config/scylla_overwrite.input.yaml"},
			Golden: "./testdata/config/scylla_overwrite.golden.yaml",
		},
		{
			Name:   "scylla overwrite multiple files",
			Input:  []string{"./testdata/config/basic.input.yaml", "./testdata/config/scylla_overwrite.input.yaml"},
			Golden: "./testdata/config/scylla_overwrite.golden.yaml",
		},
		{
			Name:   "auth token overwrite",
			Input:  []string{"./testdata/config/auth_token_overwrite.input.yaml"},
			Golden: "./testdata/config/auth_token_overwrite.golden.yaml",
		},
		{
			Name:   "https overwrite",
			Input:  []string{"./testdata/config/https_overwrite.input.yaml"},
			Golden: "./testdata/config/https_overwrite.golden.yaml",
		},
		{
			Name:   "debug overwrite",
			Input:  []string{"./testdata/config/debug_overwrite.input.yaml"},
			Golden: "./testdata/config/debug_overwrite.golden.yaml",
		},
		{
			Name:   "prometheus overwrite",
			Input:  []string{"./testdata/config/prometheus_overwrite.input.yaml"},
			Golden: "./testdata/config/prometheus_overwrite.golden.yaml",
		},
	}

	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "prometheus_port"):
			fmt.Fprint(w, 9180)
		case strings.HasSuffix(r.URL.Path, "listen_address"):
			fmt.Fprint(w, `"192.168.100.11"`)
		case strings.HasSuffix(r.URL.Path, "prometheus_address"):
			fmt.Fprint(w, `"192.168.100.11"`)
		case strings.HasSuffix(r.URL.Path, "data_file_directories"):
			fmt.Fprint(w, `["/var/lib/scylla/data"]`)
		}
	}))

	l, err := net.Listen("tcp", "127.0.0.1:10000")
	if err != nil {
		t.Skip("Failed to start test server at port 10000", err)
	}

	s.Listener = l
	s.Start()
	defer s.Close()

	for _, test := range table {
		t.Run(test.Name, func(t *testing.T) {
			c, err := parseAndValidateConfigFile(test.Input)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.enrichConfigFromAPI(context.Background(), net.JoinHostPort(c.Scylla.APIAddress, c.Scylla.APIPort)); err != nil {
				t.Fatal(err)
			}
			buf := bytes.Buffer{}
			if err := yaml.NewEncoder(&buf).Encode(c); err != nil {
				t.Fatal(err)
			}

			if *update {
				if err := ioutil.WriteFile(test.Golden, buf.Bytes(), 0666); err != nil {
					t.Error(err)
				}
			}

			golden, err := ioutil.ReadFile(test.Golden)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(buf.Bytes(), golden); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
