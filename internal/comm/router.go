// Package comm provides inter-actor communication primitives.
//
// transitional: Router is a centralized message bus. Exit condition: when Actors
// can reliably deliver messages via direct mailbox writes with equivalent
// observability guarantees, Router degrades from "required layer" to "convenience
// helper". Track via: Agent-initiated direct-send success rate ≥ 95% over 2 weeks.
package comm

import (
	"context"
	"sync"
)

// ActorStatus mirrors actor.ActorStatus to avoid import cycle.
type ActorStatus int

const (
	StatusReady ActorStatus = iota
	StatusBusy
	StatusOffline
)

// Receiver is the interface that actors must implement for the Router.
type Receiver interface {
	ID() string
	Receive(ctx context.Context, msg Message) error
	CommStatus() ActorStatus
}

// EventLogger records communication events for observability.
type EventLogger interface {
	LogCommEvent(event CommEvent)
}

// CommEvent represents a routing decision made by the Router.
type CommEvent struct {
	MessageID    string `json:"message_id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Type         string `json:"type"`
	DeliveryMode string `json:"delivery_mode"` // "direct" or "queued"
	Reason       string `json:"reason"`        // why this delivery mode was chosen
}

// Router delivers messages between Actors.
// Online+Ready actors receive directly; others get queued in mailbox.
type Router struct {
	mu      sync.RWMutex
	actors  map[string]Receiver
	mailbox *Mailbox
	logger  EventLogger
}

// NewRouter creates a router with the given mailbox backend.
func NewRouter(mailbox *Mailbox) *Router {
	return &Router{
		actors:  make(map[string]Receiver),
		mailbox: mailbox,
	}
}

// SetLogger attaches an event logger for communication observability.
func (r *Router) SetLogger(l EventLogger) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger = l
}

// Register adds an actor to the router.
func (r *Router) Register(a Receiver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actors[a.ID()] = a
}

// Unregister removes an actor from the router.
func (r *Router) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.actors, id)
}

// Send delivers a message to the target actor.
func (r *Router) Send(ctx context.Context, msg Message) error {
	r.mu.RLock()
	target, exists := r.actors[msg.To]
	logger := r.logger
	r.mu.RUnlock()

	if exists && target.CommStatus() == StatusReady {
		if logger != nil {
			logger.LogCommEvent(CommEvent{
				MessageID:    msg.ID,
				From:         msg.From,
				To:           msg.To,
				Type:         string(msg.Type),
				DeliveryMode: "direct",
				Reason:       "actor online and ready",
			})
		}
		return target.Receive(ctx, msg)
	}

	reason := "actor not registered"
	if exists {
		reason = "actor offline or busy"
	}
	if logger != nil {
		logger.LogCommEvent(CommEvent{
			MessageID:    msg.ID,
			From:         msg.From,
			To:           msg.To,
			Type:         string(msg.Type),
			DeliveryMode: "queued",
			Reason:       reason,
		})
	}
	return r.mailbox.Send(msg)
}
