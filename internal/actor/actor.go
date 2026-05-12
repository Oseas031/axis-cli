// Package actor defines the universal participant interface for Axis.
// All entities — Agents, humans, services — implement Actor.
package actor

import (
	"context"

	"github.com/axis-cli/axis/internal/comm"
)

// ActorStatus represents the availability of an Actor.
type ActorStatus int

const (
	ActorReady   ActorStatus = iota // can receive messages
	ActorBusy                       // processing, will queue
	ActorOffline                    // not available, mailbox only
)

// Actor is the universal participant interface.
type Actor interface {
	ID() string
	Receive(ctx context.Context, msg comm.Message) error
	Status() ActorStatus
}
