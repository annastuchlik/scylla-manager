// Copyright (C) 2017 ScyllaDB

package restapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/scylladb/mermaid/cluster"
	"github.com/scylladb/mermaid/uuid"
)

// ClusterService is the cluster service interface required by the repair REST
// API handlers.
type ClusterService interface {
	ListClusters(ctx context.Context, f *cluster.Filter) ([]*cluster.Cluster, error)
	GetCluster(ctx context.Context, idOrName string) (*cluster.Cluster, error)
	PutCluster(ctx context.Context, c *cluster.Cluster) error
	DeleteCluster(ctx context.Context, id uuid.UUID) error
	ListNodes(ctx context.Context, id uuid.UUID) ([]cluster.Node, error)
}

type clusterFilter struct {
	svc ClusterService
}

func (h clusterFilter) clusterCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.svc == nil {
			next.ServeHTTP(w, r)
			return
		}

		clusterID := chi.URLParam(r, "cluster_id")
		if clusterID == "" {
			respondBadRequest(w, r, errors.New("missing cluster ID"))
			return
		}

		c, err := h.svc.GetCluster(r.Context(), clusterID)
		if err != nil {
			respondError(w, r, err, fmt.Sprintf("failed to load cluster %q", clusterID))
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxClusterID, c.ID)
		ctx = context.WithValue(ctx, ctxCluster, c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type clusterHandler struct {
	clusterFilter
}

func newClusterHandler(svc ClusterService) *chi.Mux {
	m := chi.NewMux()
	h := &clusterHandler{
		clusterFilter: clusterFilter{
			svc: svc,
		},
	}

	m.Route("/clusters", func(r chi.Router) {
		r.Get("/", h.listClusters)
		r.Post("/", h.createCluster)
	})
	m.Route("/cluster/{cluster_id}", func(r chi.Router) {
		r.Use(h.clusterCtx)
		r.Get("/", h.loadCluster)
		r.Put("/", h.updateCluster)
		r.Delete("/", h.deleteCluster)
	})
	return m
}

func (h *clusterHandler) listClusters(w http.ResponseWriter, r *http.Request) {
	ids, err := h.svc.ListClusters(r.Context(), &cluster.Filter{})
	if err != nil {
		respondError(w, r, err, "failed to list clusters")
		return
	}

	if len(ids) == 0 {
		render.Respond(w, r, []struct{}{})
		return
	}
	render.Respond(w, r, ids)
}

func (h *clusterHandler) parseCluster(r *http.Request) (*cluster.Cluster, error) {
	var c cluster.Cluster
	if err := render.DecodeJSON(r.Body, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (h *clusterHandler) createCluster(w http.ResponseWriter, r *http.Request) {
	newCluster, err := h.parseCluster(r)
	if err != nil {
		respondBadRequest(w, r, err)
		return
	}
	if newCluster.ID != uuid.Nil {
		respondBadRequest(w, r, errors.Errorf("unexpected ID %q", newCluster.ID))
		return
	}

	if err := h.svc.PutCluster(r.Context(), newCluster); err != nil {
		respondError(w, r, err, "failed to create cluster")
		return
	}

	location := r.URL.ResolveReference(&url.URL{
		Path: path.Join("cluster", newCluster.ID.String()),
	})
	w.Header().Set("Location", location.String())
	w.WriteHeader(http.StatusCreated)
}

func (h *clusterHandler) loadCluster(w http.ResponseWriter, r *http.Request) {
	c := mustClusterFromCtx(r)
	render.Respond(w, r, c)
}

func (h *clusterHandler) updateCluster(w http.ResponseWriter, r *http.Request) {
	c := mustClusterFromCtx(r)

	newCluster, err := h.parseCluster(r)
	if err != nil {
		respondBadRequest(w, r, err)
		return
	}
	newCluster.ID = c.ID

	if err := h.svc.PutCluster(r.Context(), newCluster); err != nil {
		respondError(w, r, err, fmt.Sprintf("failed to update cluster %q", c.ID))
		return
	}
	render.Respond(w, r, newCluster)
}

func (h *clusterHandler) deleteCluster(w http.ResponseWriter, r *http.Request) {
	c := mustClusterFromCtx(r)

	if err := h.svc.DeleteCluster(r.Context(), c.ID); err != nil {
		respondError(w, r, err, fmt.Sprintf("failed to delete cluster %q", c.ID))
		return
	}
}
