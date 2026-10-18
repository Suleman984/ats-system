package services

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"ats-backend/config"
	"ats-backend/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Real-time chat: writes call PublishChatEvent, which goes through Postgres NOTIFY so every API
// instance hears it; each instance's listener fans events out to its connected SSE streams.

const chatChannel = "ats_chat"

// ChatEvent is what streams receive. Message is set for "message" events; Reader for "read" events.
type ChatEvent struct {
	Type          string          `json:"type"` // "message" or "read"
	ApplicationID uuid.UUID       `json:"application_id"`
	CompanyID     uuid.UUID       `json:"company_id"`
	MessageID     *uuid.UUID      `json:"message_id,omitempty"`
	Reader        string          `json:"reader,omitempty"`
	Message       *models.Message `json:"message,omitempty"`
}

// PublishChatEvent announces an event; failures only delay delivery until the next reload.
func PublishChatEvent(ev ChatEvent) {
	ev.Message = nil // NOTIFY payloads are capped at 8 kB; listeners load the message by ID.
	payload, err := json.Marshal(ev)
	if err != nil {
		return
	}
	if err := config.DB.Exec("select pg_notify(?, ?)", chatChannel, string(payload)).Error; err != nil {
		log.Printf("chat notify: %v", err)
	}
}

type chatSubscriber struct {
	companyID     uuid.UUID
	applicationID *uuid.UUID // set for candidate streams, which only see their own thread
	events        chan ChatEvent
}

var chatHub = struct {
	sync.RWMutex
	subs map[*chatSubscriber]struct{}
}{subs: map[*chatSubscriber]struct{}{}}

// SubscribeChat registers a stream. Pass applicationID to limit it to one conversation.
func SubscribeChat(companyID uuid.UUID, applicationID *uuid.UUID) (<-chan ChatEvent, func()) {
	sub := &chatSubscriber{companyID: companyID, applicationID: applicationID, events: make(chan ChatEvent, 16)}
	chatHub.Lock()
	chatHub.subs[sub] = struct{}{}
	chatHub.Unlock()
	return sub.events, func() {
		chatHub.Lock()
		delete(chatHub.subs, sub)
		chatHub.Unlock()
	}
}

func dispatchChatEvent(ev ChatEvent) {
	chatHub.RLock()
	defer chatHub.RUnlock()
	for sub := range chatHub.subs {
		if sub.companyID != ev.CompanyID || (sub.applicationID != nil && *sub.applicationID != ev.ApplicationID) {
			continue
		}
		select {
		case sub.events <- ev:
		default: // A stuck client must not block everyone; it catches up on reconnect.
		}
	}
}

func hasChatSubscribers(companyID uuid.UUID) bool {
	chatHub.RLock()
	defer chatHub.RUnlock()
	for sub := range chatHub.subs {
		if sub.companyID == companyID {
			return true
		}
	}
	return false
}

// StartChatListener keeps a dedicated LISTEN connection open, reconnecting with backoff.
// Needs a direct or session-pooler connection; Supabase's transaction pooler (6543) can't LISTEN.
func StartChatListener() {
	go func() {
		backoff := time.Second
		for {
			if err := listenForChatEvents(); err != nil {
				log.Printf("chat listener: %v (retrying in %s)", err, backoff)
			}
			time.Sleep(backoff)
			backoff = min(backoff*2, 30*time.Second)
		}
	}()
}

func listenForChatEvents() error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, "listen "+chatChannel); err != nil {
		return err
	}
	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		var ev ChatEvent
		if json.Unmarshal([]byte(n.Payload), &ev) != nil || !hasChatSubscribers(ev.CompanyID) {
			continue
		}
		if ev.Type == "message" && ev.MessageID != nil {
			var msg models.Message
			if err := config.DB.First(&msg, "id = ?", *ev.MessageID).Error; err != nil {
				continue
			}
			ev.Message = &msg
		}
		dispatchChatEvent(ev)
	}
}
