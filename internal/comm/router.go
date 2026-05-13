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

// Router delivers messages between Actors.
// Online+Ready actors receive directly; others get queued in mailbox.
type Router struct {
	mu      sync.RWMutex
	actors  map[string]Receiver
	mailbox *Mailbox
}

// NewRouter creates a router with the given mailbox backend.
func NewRouter(mailbox *Mailbox) *Router {
	return &Router{
		actors:  make(map[string]Receiver),
		mailbox: mailbox,
	}
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
	r.mu.RUnlock()

	if exists && target.CommStatus() == StatusReady {
		return target.Receive(ctx, msg)
	}
	return r.mailbox.Send(msg)
}
