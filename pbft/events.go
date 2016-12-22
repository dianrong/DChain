// Copyright Dianrong.com Corp. 2016 All Rights Reserved.
//
// The Roc is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package pbft

import "github.com/op/go-logging"

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/util/events")
}

type Event interface{}

// Receiver is a consumer of events, ProcessEvent will be called serially
// as events arrive
type Receiver interface {
	// ProcessEvent delivers an event to the Receiver, if it returns non-nil, the return is the next processed event
	ProcessEvent(e Event) Event
}

// threaded holds an exit channel to allow threads to break from a select
type threaded struct {
	exit chan struct{}
}

// halt tells the threaded object's thread to exit
func (t *threaded) Halt() {
	select {
	case <-t.exit:
		logger.Warning("Attempted to halt a threaded object twice")
	default:
		close(t.exit)
	}
}

// ------------------------------------------------------------
//
// Event Manager
//
// ------------------------------------------------------------

// Manager provides a serialized interface for submitting events to
// a Receiver on the other side of the queue
type Manager interface {
	Inject(Event)         // A temporary interface to allow the event manager thread to skip the queue
	Queue() chan<- Event  // Get a write-only reference to the queue, to submit events
	SetReceiver(Receiver) // Set the target to route events to
	Start()               // Starts the Manager thread TODO, these thread management things should probably go away
	Halt()                // Stops the Manager thread
}

// managerImpl is an implementation of Manger
type managerImpl struct {
	threaded
	receiver Receiver
	events   chan Event
}

// NewManagerImpl creates an instance of managerImpl
func NewManagerImpl() Manager {
	return &managerImpl{
		events:   make(chan Event),
		threaded: threaded{make(chan struct{})},
	}
}

// SetReceiver sets the destination for events
func (em *managerImpl) SetReceiver(receiver Receiver) {
	em.receiver = receiver
}

// Start creates the go routine necessary to deliver events
func (em *managerImpl) Start() {
	go em.eventLoop()
}

// queue returns a write only reference to the event queue
func (em *managerImpl) Queue() chan<- Event {
	return em.events
}

// SendEvent performs the event loop on a receiver to completion
func SendEvent(receiver Receiver, event Event) {
	next := event
	for {
		// If an event returns something non-nil, then process it as a new event
		next = receiver.ProcessEvent(next)
		if next == nil {
			break
		}
	}
}

// Inject can only safely be called by the managerImpl thread itself, it skips the queue
func (em *managerImpl) Inject(event Event) {
	if em.receiver != nil {
		SendEvent(em.receiver, event)
	}
}

// eventLoop is where the event thread loops, delivering events
func (em *managerImpl) eventLoop() {
	for {
		select {
		case next := <-em.events:
			em.Inject(next)
		case <-em.exit:
			logger.Debug("eventLoop told to exit")
			return
		}
	}
}
