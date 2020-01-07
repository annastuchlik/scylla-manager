// Copyright (C) 2017 ScyllaDB

package backup

import (
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/scylladb/go-set/strset"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/mermaid/pkg/util/uuid"
)

// ListFilter specifies filtering for backup listing.
type ListFilter struct {
	ClusterID   uuid.UUID `json:"cluster_id"`
	Keyspace    []string  `json:"keyspace"`
	SnapshotTag string    `json:"snapshot_tag"`
	MinDate     time.Time `json:"min_date"`
	MaxDate     time.Time `json:"max_date"`
}

// ListItem represents contents of a snapshot within list boundaries.
type ListItem struct {
	ClusterID    uuid.UUID `json:"cluster_id"`
	Units        []Unit    `json:"units,omitempty"`
	SnapshotTags []string  `json:"snapshot_tags"`

	unitsHash uint64 // Internal usage only
}

// FilesInfo specifies paths to files backed up for a table (and node) within
// a location.
// Note that a backup for a table usually consists of multiple instances of
// FilesInfo since data is replicated across many nodes.
type FilesInfo struct {
	Keyspace string   `json:"keyspace"`
	Table    string   `json:"table"`
	Location Location `json:"location"`
	Manifest string   `json:"manifest"`
	SST      string   `json:"sst"`
	Files    []string `json:"files"`
}

func makeFilesInfo(m remoteManifest) FilesInfo {
	t := FilesInfo{
		Keyspace: m.Keyspace,
		Table:    m.Table,
		Location: m.Location,
		Manifest: m.RemoteManifestFile(),
		SST:      m.RemoteSSTableVersionDir(),
		Files:    m.FilesExpanded,
	}

	return t
}

// Target specifies what should be backed up and where.
type Target struct {
	Units            []Unit     `json:"units,omitempty"`
	DC               []string   `json:"dc,omitempty"`
	Location         []Location `json:"location"`
	Retention        int        `json:"retention"`
	RateLimit        []DCLimit  `json:"rate_limit"`
	SnapshotParallel []DCLimit  `json:"snapshot_parallel"`
	UploadParallel   []DCLimit  `json:"upload_parallel"`
	Continue         bool       `json:"continue"`
}

// Unit represents keyspace and its tables.
type Unit struct {
	Keyspace  string   `json:"keyspace" db:"keyspace_name"`
	Tables    []string `json:"tables,omitempty"`
	AllTables bool     `json:"all_tables"`
}

func (u Unit) MarshalUDT(name string, info gocql.TypeInfo) ([]byte, error) {
	f := gocqlx.DefaultMapper.FieldByName(reflect.ValueOf(u), name)
	return gocql.Marshal(info, f.Interface())
}

func (u *Unit) UnmarshalUDT(name string, info gocql.TypeInfo, data []byte) error {
	f := gocqlx.DefaultMapper.FieldByName(reflect.ValueOf(u), name)
	return gocql.Unmarshal(info, data, f.Addr().Interface())
}

// Run tracks backup progress, shares ID with scheduler.Run that initiated it.
type Run struct {
	ClusterID uuid.UUID
	TaskID    uuid.UUID
	ID        uuid.UUID

	PrevID      uuid.UUID
	SnapshotTag string
	Units       []Unit
	DC          []string
	Location    []Location
	StartTime   time.Time
	Done        bool

	clusterName string
}

// RunProgress describes backup progress on per file basis.
//
// Each RunProgress either has Uploaded or Skipped fields set to respective
// amount of bytes. Failed shows amount of bytes that is assumed to have
// failed. Since current implementation doesn't support resume at file level
// this value will always be the same as Uploaded as file needs to be uploaded
// again. In summary Failed is supposed to mean, out of uploaded bytes how much
// bytes have to be uploaded again.
type RunProgress struct {
	ClusterID  uuid.UUID
	TaskID     uuid.UUID
	RunID      uuid.UUID
	AgentJobID int64

	Host      string
	Unit      int64
	TableName string

	Files       []string
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       string
	Size        int64 // Total file size in bytes.
	Uploaded    int64 // Amount of total uploaded bytes.
	Skipped     int64 // Amount of skipped bytes because file was present.
	// Amount of bytes that have been uploaded but due to error have to be
	// uploaded again.
	Failed int64
}

type progress struct {
	Size        int64      `json:"size"`
	Uploaded    int64      `json:"uploaded"`
	Skipped     int64      `json:"skipped"`
	Failed      int64      `json:"failed"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// Progress groups uploading progress for all backed up hosts.
type Progress struct {
	progress
	SnapshotTag string         `json:"snapshot_tag"`
	DC          []string       `json:"dcs,omitempty"`
	Hosts       []HostProgress `json:"hosts,omitempty"`
}

// HostProgress groups uploading progress for keyspaces belonging to this host.
type HostProgress struct {
	progress

	Host      string             `json:"host"`
	Keyspaces []KeyspaceProgress `json:"keyspaces,omitempty"`
}

// KeyspaceProgress groups uploading progress for the tables belonging to this
// keyspace.
type KeyspaceProgress struct {
	progress

	Keyspace string          `json:"keyspace"`
	Tables   []TableProgress `json:"tables,omitempty"`
}

// TableProgress defines progress for the table.
type TableProgress struct {
	progress

	Table string `json:"table"`
	Error string `json:"error,omitempty"`
}

// Provider specifies type of remote storage like S3 etc.
type Provider string

// Provider enumeration.
const (
	S3 = Provider("s3")
)

func (p Provider) String() string {
	return string(p)
}

// MarshalText implements encoding.TextMarshaler.
func (p Provider) MarshalText() (text []byte, err error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *Provider) UnmarshalText(text []byte) error {
	if s := string(text); !providers.Has(s) {
		return errors.Errorf("unrecognised provider %q", text)
	}
	*p = Provider(text)
	return nil
}

var providers = strset.New(S3.String())

// Location specifies storage provider and container/resource for a DC.
type Location struct {
	DC       string   `json:"dc"`
	Provider Provider `json:"provider"`
	Path     string   `json:"path"`
}

func (l Location) String() string {
	p := l.Provider.String() + ":" + l.Path
	if l.DC != "" {
		p = l.DC + ":" + p
	}
	return p
}

func (l Location) MarshalText() (text []byte, err error) {
	return []byte(l.String()), nil
}

func (l *Location) UnmarshalText(text []byte) error {
	// Providers require that resource names are DNS compliant.
	// The following is a super simplified DNS (plus provider prefix)
	// matching regexp.
	pattern := regexp.MustCompile(`^(([a-zA-Z0-9\-\_\.]+):)?([a-z0-9]+):([a-z0-9\-\.]+)$`)

	m := pattern.FindSubmatch(text)
	if m == nil {
		return errors.Errorf("invalid location %q, the format is [dc:]<provider>:<path> ex. s3:my-bucket, the path must be DNS compliant", string(text))
	}

	if err := l.Provider.UnmarshalText(m[3]); err != nil {
		return errors.Wrapf(err, "invalid location %q", string(text))
	}

	l.DC = string(m[2])
	l.Path = string(m[4])

	return nil
}

func (l Location) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	return l.MarshalText()
}

func (l *Location) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	return l.UnmarshalText(data)
}

// RemoteName returns the rclone remote name for that location.
func (l Location) RemoteName() string {
	return l.Provider.String()
}

// RemotePath returns string that can be used with rclone to specify a path in
// the given location.
func (l Location) RemotePath(p string) string {
	r := l.RemoteName()
	if r != "" {
		r += ":"
	}
	return path.Join(r+l.Path, p)
}

// DCLimit specifies a rate limit for a DC.
type DCLimit struct {
	DC    string `json:"dc"`
	Limit int    `json:"limit"`
}

func (l DCLimit) String() string {
	p := fmt.Sprint(l.Limit)
	if l.DC != "" {
		p = l.DC + ":" + p
	}
	return p
}

func (l DCLimit) MarshalText() (text []byte, err error) {
	return []byte(l.String()), nil
}

func (l *DCLimit) UnmarshalText(text []byte) error {
	pattern := regexp.MustCompile(`^(([a-zA-Z0-9\-\_\.]+):)?([0-9]+)$`)

	m := pattern.FindSubmatch(text)
	if m == nil {
		return errors.Errorf("invalid limit %q, the format is [dc:]<number>", string(text))
	}

	limit, err := strconv.ParseInt(string(m[3]), 10, 64)
	if err != nil {
		return errors.Wrap(err, "invalid limit value")
	}

	l.DC = string(m[2])
	l.Limit = int(limit)

	return nil
}

func dcLimitDCAtPos(s []DCLimit) func(int) (string, string) {
	return func(i int) (string, string) {
		return s[i].DC, s[i].String()
	}
}

// taskProperties is the main data structure of the runner.Properties blob.
type taskProperties struct {
	Keyspace         []string   `json:"keyspace"`
	DC               []string   `json:"dc"`
	Location         []Location `json:"location"`
	Retention        int        `json:"retention"`
	RateLimit        []DCLimit  `json:"rate_limit"`
	SnapshotParallel []DCLimit  `json:"snapshot_parallel"`
	UploadParallel   []DCLimit  `json:"upload_parallel"`
	Continue         bool       `json:"continue"`
}

func defaultTaskProperties() taskProperties {
	return taskProperties{
		Retention: 3,
		Continue:  true,
	}
}
