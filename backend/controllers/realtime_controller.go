package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ats-backend/config"
	"ats-backend/middleware"
	"ats-backend/models"
	"ats-backend/services"
	"ats-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Server-Sent Events: the browser holds one stream open and receives chat events as they happen.
// Sending still goes through the normal POST endpoints.

const streamHeartbeat = 25 * time.Second

// ChatEvents streams every chat event for the signed-in admin's company (one stream per tab
// serves both the open conversation and the unread bell).
func ChatEvents(c *gin.Context) {
	events, unsubscribe := services.SubscribeChat(companyID(c), nil)
	defer unsubscribe()
	streamChatEvents(c, events, func() bool { return middleware.SessionStillActive(c) }, nil)
}

// CandidateStreamToken trades the candidate's email + application ID for a short-lived token, so
// those credentials never appear in the stream URL (EventSource can only send GET).
func CandidateStreamToken(c *gin.Context) {
	var req struct {
		Email         string    `json:"email" binding:"required,email"`
		ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	app, ok := findCandidateApplication(c, req.ApplicationID, req.Email)
	if !ok {
		return
	}
	token, err := utils.GenerateCandidateStreamToken(app.ID, app.CompanyID)
	if err != nil {
		serverError(c, "Failed to start live updates")
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// CandidateChatEvents streams one application's conversation to the candidate.
func CandidateChatEvents(c *gin.Context) {
	applicationID, company, err := utils.VerifyCandidateStreamToken(c.Query("token"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Stream token is invalid or expired"})
		return
	}
	stillExists := func() bool {
		var n int64
		config.DB.Model(&models.Application{}).Where("id = ?", applicationID).Count(&n)
		return n == 1
	}
	if !stillExists() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}
	events, unsubscribe := services.SubscribeChat(company, &applicationID)
	defer unsubscribe()
	streamChatEvents(c, events, stillExists, func(ev services.ChatEvent) services.ChatEvent {
		// Candidates don't learn which recruiter wrote a message.
		if ev.Message != nil && ev.Message.SenderType == models.SenderRecruiter {
			redacted := *ev.Message
			redacted.SenderID, redacted.SenderEmail = nil, ""
			ev.Message = &redacted
		}
		ev.CompanyID = uuid.Nil
		return ev
	})
}

// streamChatEvents writes events until the client leaves or stillAllowed turns false.
func streamChatEvents(c *gin.Context, events <-chan services.ChatEvent, stillAllowed func() bool,
	transform func(services.ChatEvent) services.ChatEvent) {
	// The server's WriteTimeout would cut the stream after a minute; lift it for this request only.
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})
	h := c.Writer.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-store")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no") // stop nginx-style proxies from buffering events
	c.Status(http.StatusOK)

	// retry: how long the browser waits before reconnecting after a drop.
	fmt.Fprint(c.Writer, "retry: 3000\n\n")
	c.Writer.Flush()

	heartbeat := time.NewTicker(streamHeartbeat)
	defer heartbeat.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			if !stillAllowed() {
				return
			}
			// Comment lines keep proxies and load balancers from closing an idle connection.
			fmt.Fprint(c.Writer, ": ping\n\n")
			c.Writer.Flush()
		case ev := <-events:
			if transform != nil {
				ev = transform(ev)
			}
			data, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", ev.Type, data)
			c.Writer.Flush()
		}
	}
}
