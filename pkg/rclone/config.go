// Copyright (C) 2017 ScyllaDB

package rclone

import (
	"fmt"
	"sync"

	"github.com/rclone/rclone/fs"
	"github.com/scylladb/mermaid/pkg"
)

// InitFsConfig enables in-memory config and sets default config values
// expected for correct agent behaviour.
func InitFsConfig() {
	initInMemoryConfig()

	// Don't use JSON log format in logging.
	fs.Config.UseJSONLog = false
	// Pass all logs, our logger decides which one to print.
	fs.Config.LogLevel = fs.LogLevelDebug
	// Delete even if there are I/O errors.
	fs.Config.IgnoreErrors = true
	// Do not compare hash post upload, prevents from calculating hashes in
	// rclone versions >= 1.48.
	fs.Config.IgnoreChecksum = true
	// Only use size to compare files.
	fs.Config.SizeOnly = true
	// Don't update destination mod-time if files identical.
	fs.Config.NoUpdateModTime = true
	// Set proper agent for backend clients.
	fs.Config.UserAgent = UserAgent()
	// How many times to retry low level operations like copy file.
	fs.Config.LowLevelRetries = 5
	// How many stat groups to keep in memory.
	fs.Config.MaxStatsGroups = 1000
}

func initInMemoryConfig() {
	c := new(inMemoryConf)
	fs.ConfigFileGet = c.Get
	fs.ConfigFileSet = c.Set
	fs.Infof(nil, "registered in-memory fs config")
}

// inMemoryConf is in-memory implementation of rclone configuration for
// remote file systems.
type inMemoryConf struct {
	mu       sync.Mutex
	sections map[string]map[string]string
}

// Get config key under section returning the the value and true if found or
// ("", false) otherwise.
func (c *inMemoryConf) Get(section, key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sections == nil {
		return "", false
	}
	s, ok := c.sections[section]
	if !ok {
		return "", false
	}
	v, ok := s[key]
	return v, ok
}

// Set the key in section to value.
// It doesn't save the config file.
func (c *inMemoryConf) Set(section, key, value string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sections == nil {
		c.sections = make(map[string]map[string]string)
	}
	s, ok := c.sections[section]
	if !ok {
		s = make(map[string]string)
	}
	if value == "" {
		delete(c.sections[section], value)
	} else {
		s[key] = value
		c.sections[section] = s
	}
	return
}

// UserAgent returns string value that can be used as identifier in client
// calls to the service providers.
func UserAgent() string {
	return fmt.Sprintf("Scylla Manager Agent %s", pkg.Version())
}
