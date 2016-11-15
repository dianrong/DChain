package pbft


// commitedEvent is sent when a requested commit completes
type committedEvent struct {
	tag    interface{}
}

// rolledBackEvent is sent when a requested rollback completes
type rolledBackEvent struct{}

type externalEventReceiver struct {
	manager Manager
}

// RecvMsg is called by the stack when a new message is received
func (eer *externalEventReceiver) RecvMsg() error {
	eer.manager.Queue() <- batchMessageEvent{

	}
	return nil
}

// Committed is called whenever Commit completes, no-op for noops as it uses the legacy synchronous api
func (eer *externalEventReceiver) Committed(tag interface{}, target *pb.BlockchainInfo) {
	eer.manager.Queue() <- committedEvent{tag, target}
}

// RolledBack is called whenever a Rollback completes, no-op for noops as it uses the legacy synchronous api
func (eer *externalEventReceiver) RolledBack(tag interface{}) {
	eer.manager.Queue() <- rolledBackEvent{}
}
