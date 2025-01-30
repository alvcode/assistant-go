package handler

import (
	"net/http"
)

type HeartbeatHandler struct {
}

func NewHeartbeatHandler() *HeartbeatHandler {
	return &HeartbeatHandler{}
}

// Heartbeat
// @Summary Heartbeat metric
// @Tags Metrics
// @Success 204
// @Failure 400
// @Router /api/heartbeat [get]
func (h *HeartbeatHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
