package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"taskflow-shivam-goyal/backend/internal/middleware"
	"taskflow-shivam-goyal/backend/internal/realtime"
	"taskflow-shivam-goyal/backend/internal/response"
)

const sseKeepAliveInterval = 30 * time.Second

func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_projects_events_unauthorized_response_failed", "error", err)
		}
		return
	}

	projectID := chi.URLParam(r, "id")
	if err := h.service.AuthorizeAccess(r.Context(), currentUserID, projectID); err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			if writeErr := response.NotFound(w, "project not found"); writeErr != nil {
				h.logger.Error("http_projects_events_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_projects_events_forbidden_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_projects_events_access_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_projects_events_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logger.Error("http_projects_events_streaming_not_supported")
		if err := response.InternalServerError(w); err != nil {
			h.logger.Error("http_projects_events_streaming_response_failed", "error", err)
		}
		return
	}

	if h.events == nil {
		h.logger.Error("http_projects_events_manager_missing")
		if err := response.InternalServerError(w); err != nil {
			h.logger.Error("http_projects_events_manager_response_failed", "error", err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := h.events.Subscribe(projectID)
	defer unsubscribe()

	if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
		h.logger.Warn("http_projects_events_connected_write_failed", "error", err)
		return
	}
	flusher.Flush()

	keepAliveTicker := time.NewTicker(sseKeepAliveInterval)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			h.logger.Info("http_projects_events_client_disconnected", "project_id", projectID)
			return
		case event, ok := <-events:
			if !ok {
				return
			}

			if err := writeSSEEvent(w, event); err != nil {
				h.logger.Warn("http_projects_events_write_failed", "project_id", projectID, "error", err)
				return
			}
			flusher.Flush()
		case <-keepAliveTicker.C:
			if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
				h.logger.Warn("http_projects_events_keepalive_failed", "project_id", projectID, "error", err)
				return
			}
			flusher.Flush()
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, event realtime.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal sse event: %w", err)
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", event.Type); err != nil {
		return fmt.Errorf("write sse event name: %w", err)
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return fmt.Errorf("write sse event payload: %w", err)
	}

	return nil
}
