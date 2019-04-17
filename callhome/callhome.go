// Copyright (C) 2017 ScyllaDB

package callhome

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-version"
	"github.com/scylladb/go-log"
	"github.com/scylladb/mermaid"
	"github.com/scylladb/mermaid/internal/osutil"
	"github.com/scylladb/mermaid/uuid"
)

const (
	statusInstall       = "mi"
	statusDaily         = "md"
	statusInstallDocker = "mdi"
	statusDailyDocker   = "mdd"
)

const (
	defaultHostURL = "https://repositories.scylladb.com/scylla/check_version"
)

// Checker is a container for all dependencies needed for making calls to
// check manager version.
// Checker will make HTTP GET request to the HostURL with parameters extracted
// from Env and Version.
// If version returned in the response is higher than the current running
// version Checker will add info level entry to the Logger.
type Checker struct {
	HostURL string
	Version string
	Client  *http.Client
	Env     OSEnv
	Logger  log.Logger
}

// OSEnv contains all methods required by the Checker.
type OSEnv interface {
	MacUUID() uuid.UUID
	RegUUID() uuid.UUID
	LinuxDistro() string
	Docker() bool
}

type osenv struct{}

func (e osenv) MacUUID() uuid.UUID {
	return osutil.MacUUID()
}

func (e osenv) RegUUID() uuid.UUID {
	return osutil.RegUUID()
}

func (e osenv) LinuxDistro() string {
	return string(osutil.LinuxDistro())
}

func (e osenv) Docker() bool {
	return osutil.Docker()
}

// DefaultEnv represents default running environment.
var DefaultEnv osenv

// NewChecker creates new service.
func NewChecker(hostURL, version string, l log.Logger, env OSEnv) *Checker {
	if hostURL == "" {
		hostURL = defaultHostURL
	}
	if version == "" {
		version = mermaid.Version()
	}

	return &Checker{
		HostURL: hostURL,
		Version: version,
		Logger:  l,
		Client:  http.DefaultClient,
		Env:     env,
	}
}

type checkResponse struct {
	LatestPatchVersion string `json:"latest_patch_version"`
	Version            string `json:"version"`
}

// CheckForUpdates sends request for comparing current version with installed.
// If install is true it sends install status.
func (s *Checker) CheckForUpdates(ctx context.Context, install bool) error {
	u, err := url.Parse(s.HostURL)
	if err != nil {
		panic(err.Error())
	}
	q := u.Query()
	q.Add("system", "scylla-manager")
	q.Add("version", s.Version)
	q.Add("uu", s.Env.MacUUID().String())
	q.Add("rid", s.Env.RegUUID().String())
	q.Add("rtype", s.Env.LinuxDistro())
	sts := statusDaily
	docker := s.Env.Docker()
	if docker {
		sts = statusDailyDocker
	}
	if install {
		if docker {
			sts = statusInstallDocker
		} else {
			sts = statusInstall
		}
	}
	q.Add("sts", sts)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	check := checkResponse{}
	if err := json.Unmarshal(d, &check); err != nil {
		return err
	}

	available, err := version.NewVersion(check.Version)
	if err != nil {
		return err
	}
	installed, err := version.NewVersion(s.Version)
	if err != nil {
		return err
	}
	if installed.LessThan(available) {
		s.Logger.Info(ctx, "New Scylla Manager version is available",
			"installed", s.Version, "available", check.Version)
	}
	return nil
}
