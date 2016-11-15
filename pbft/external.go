package pbft


// stateUpdatedEvent is sent when state transfer completes
type stateUpdatedEvent struct {
	chkpt  *checkpointMessage
	target *pb.BlockchainInfo
}

// executedEvent is sent when a requested execution completes
type executedEvent struct {
	tag interface{}
}

// commitedEvent is sent when a requested commit completes
type committedEvent struct {
	tag    interface{}
	target *pb.BlockchainInfo
}

// rolledBackEvent is sent when a requested rollback completes
type rolledBackEvent struct{}

type externalEventReceiver struct {
	manager events.Manager
}

// RecvMsg is called by the stack when a new message is received
func (eer *externalEventReceiver) RecvMsg(ocMsg *pb.Message, senderHandle *pb.PeerID) error {
	eer.manager.Queue() <- batchMessageEvent{
		msg:    ocMsg,
		sender: senderHandle,
	}
	return nil
}

// Executed is called whenever Execute completes, no-op for noops as it uses the legacy synchronous api
func (eer *externalEventReceiver) Executed(tag interface{}) {
	eer.manager.Queue() <- executedEvent{tag}
}

// Committed is called whenever Commit completes, no-op for noops as it uses the legacy synchronous api
func (eer *externalEventReceiver) Committed(tag interface{}, target *pb.BlockchainInfo) {
	eer.manager.Queue() <- committedEvent{tag, target}
}

// RolledBack is called whenever a Rollback completes, no-op for noops as it uses the legacy synchronous api
func (eer *externalEventReceiver) RolledBack(tag interface{}) {
	eer.manager.Queue() <- rolledBackEvent{}
}

// StateUpdated is a signal from the stack that it has fast-forwarded its state
func (eer *externalEventReceiver) StateUpdated(tag interface{}, target *pb.BlockchainInfo) {
	eer.manager.Queue() <- stateUpdatedEvent{
		chkpt:  tag.(*checkpointMessage),
		target: target,
	}
}