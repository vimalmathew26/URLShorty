package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"urlshorty/internal/core"
)

type Handlers struct {
	svc     *core.Service
	baseURL string
}

func NewHandlers(svc *core.Service, baseURL string) *Handlers {
	return &Handlers{svc: svc, baseURL: baseURL}
}

// ---- endpoints ----

func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handlers) Shorten(c *gin.Context) {
	var in core.CreateRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		jsonError(c, http.StatusBadRequest, "invalid json body")
		return
	}
	rec, err := h.svc.Shorten(c.Request.Context(), in)
	if err != nil {
		switch err {
		case core.ErrInvalidURL, core.ErrInvalidCode:
			jsonError(c, http.StatusBadRequest, err.Error())
		case core.ErrConflict:
			jsonError(c, http.StatusConflict, err.Error())
		default:
			jsonError(c, http.StatusInternalServerError, "internal error")
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":      rec.Code,
		"short_url": h.baseURL + "/" + rec.Code,
	})
}

func (h *Handlers) Redirect(c *gin.Context) {
	code := c.Param("code")
	rec, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		switch err {
		case core.ErrInvalidCode:
			jsonError(c, http.StatusBadRequest, err.Error())
		case core.ErrNotFound:
			jsonError(c, http.StatusNotFound, "not found")
		case core.ErrExpired:
			jsonError(c, http.StatusGone, "link expired")
		default:
			jsonError(c, http.StatusInternalServerError, "internal error")
		}
		return
	}

	// Best-effort hit counting (async) with a proper context.
	go func(code string) {
		_ = h.svc.RecordHit(context.Background(), code)
	}(code)

	c.Redirect(http.StatusMovedPermanently, rec.LongURL)
}

func (h *Handlers) Metadata(c *gin.Context) {
	code := c.Param("code")
	rec, err := h.svc.Metadata(c.Request.Context(), code)
	if err != nil {
		switch err {
		case core.ErrInvalidCode:
			jsonError(c, http.StatusBadRequest, err.Error())
		case core.ErrNotFound:
			jsonError(c, http.StatusNotFound, "not found")
		default:
			jsonError(c, http.StatusInternalServerError, "internal error")
		}
		return
	}
	expired := false
	if rec.ExpiresAt != nil && time.Now().After(*rec.ExpiresAt) {
		expired = true
	}
	c.JSON(http.StatusOK, gin.H{
		"code":       rec.Code,
		"url":        rec.LongURL,
		"created_at": rec.CreatedAt,
		"expires_at": rec.ExpiresAt,
		"hits":       rec.Hits,
		"expired":    expired,
		"short_url":  h.baseURL + "/" + rec.Code,
	})
}

// ---- helpers ----

func jsonError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
}
