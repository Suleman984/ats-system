package services

import (
	"testing"

	"github.com/google/uuid"
)

func received(ch <-chan ChatEvent) int {
	n := 0
	for {
		select {
		case <-ch:
			n++
		default:
			return n
		}
	}
}

func TestChatEventsStayWithinCompanyAndConversation(t *testing.T) {
	companyA, companyB := uuid.New(), uuid.New()
	app1, app2 := uuid.New(), uuid.New()

	recruiterA, stopA := SubscribeChat(companyA, nil)
	defer stopA()
	recruiterB, stopB := SubscribeChat(companyB, nil)
	defer stopB()
	candidate1, stopC := SubscribeChat(companyA, &app1)
	defer stopC()

	dispatchChatEvent(ChatEvent{Type: "message", CompanyID: companyA, ApplicationID: app1})
	dispatchChatEvent(ChatEvent{Type: "message", CompanyID: companyA, ApplicationID: app2})

	if got := received(recruiterA); got != 2 {
		t.Errorf("company A recruiter got %d events, want 2", got)
	}
	if got := received(recruiterB); got != 0 {
		t.Errorf("company B recruiter got %d events from company A, want 0", got)
	}
	if got := received(candidate1); got != 1 {
		t.Errorf("candidate for app1 got %d events, want only their own 1", got)
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	company := uuid.New()
	ch, stop := SubscribeChat(company, nil)
	stop()
	dispatchChatEvent(ChatEvent{Type: "read", CompanyID: company})
	if received(ch) != 0 || hasChatSubscribers(company) {
		t.Fatal("unsubscribed stream still receives events")
	}
}
