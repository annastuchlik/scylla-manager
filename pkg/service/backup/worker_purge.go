// Copyright (C) 2017 ScyllaDB

package backup

import (
	"context"
	"time"

	"github.com/scylladb/mermaid/pkg/util/parallel"
	"github.com/scylladb/mermaid/pkg/util/timeutc"
)

func (w *worker) Purge(ctx context.Context, hosts []hostInfo, policy int) (err error) {
	w.Logger.Info(ctx, "Purging old data...")
	defer func(start time.Time) {
		if err != nil {
			w.Logger.Error(ctx, "Purging old data failed see exact errors above", "duration", timeutc.Since(start))
		} else {
			w.Logger.Info(ctx, "Done purging old data", "duration", timeutc.Since(start))
		}
	}(timeutc.Now())

	return hostsInParallel(hosts, parallel.NoLimit, func(h hostInfo) error {
		w.Logger.Info(ctx, "Purging old data on host", "host", h.IP)
		err := w.purgeHost(ctx, h, policy)
		if err != nil {
			w.Logger.Error(ctx, "Purging old data failed on host", "host", h.IP, "error", err)
		} else {
			w.Logger.Info(ctx, "Done purging old data on host", "host", h.IP)
		}
		return err
	})
}

func (w *worker) purgeHost(ctx context.Context, h hostInfo, policy int) error {
	if err := w.Client.DeleteSnapshot(ctx, h.IP, w.SnapshotTag); err != nil {
		w.Logger.Error(ctx, "Failed to delete uploaded snapshot",
			"host", h.IP,
			"tag", w.SnapshotTag,
			"error", err,
		)
	} else {
		w.Logger.Info(ctx, "Deleted uploaded snapshot",
			"host", h.IP,
			"tag", w.SnapshotTag,
		)
	}

	p := &purger{
		ClusterID:      w.ClusterID,
		TaskID:         w.TaskID,
		HostInfo:       h,
		Policy:         policy,
		Client:         w.Client,
		ManifestHelper: newPurgerManifestHelper(h.IP, h.Location, w.Client, w.Logger),
		Logger:         w.Logger,
	}

	if err := p.purge(ctx); err != nil {
		w.Logger.Error(ctx, "Failed to delete remote stale snapshots",
			"host", h.IP,
			"location", h.Location,
			"error", err,
		)
		return err
	}

	return nil
}
