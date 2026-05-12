package comm

import (
	"context"
	"sync"

	"github.com/axis-cli/axis/internal/actor"
)

// Router delivers messages between Actors.
// Online+Ready actors receive directly; others get queued in mailbox.
type Router struct {
	mu       sync.RWMutex
	actors   map[string]actor.Actor
	mailbox  *Mailbox
}

// NewRouter creates a router with the given mailbox backend.
func NewRouter(mailbox *Mailbox) *Router {
	return &Router{
		actors:  make(map[string]actor.Actor),
		mailbox: mailbox,
	}
}

// Register adds an actor to the router.
func (r *Router) Register(a actor.Actor) {
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
// If the actor is online and ready, delivers directly.
// Otherwise queues in mailbox.
func (r *Router) Send(ctx context.Context, msg actor.Message) error {
	r.mu.RLock()
	target, exists := r.actors[msg.To]
	r.mu.RUnlock()

	if exists && target.Status() == actor.ActorReady {
		return target.Receive(ctx, msg)
	}
	// Queue in mailbox for later retrieval
	return r.mailbox.Send(msg)
}
