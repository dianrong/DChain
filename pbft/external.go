package pbft

import "github.com/ethereum/go-ethereum/core/types"


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
func (eer *externalEventReceiver) RecvMsg(msg *types.Message) error {
	logger.Infof("RecvMsg external.go")

	eer.manager.Queue() <- types.BatchMessageEvent{
		Msg: msg,
	}
	return nil
}

// RolledBack is called whenever a Rollback completes, no-op for noops as it uses the legacy synchronous api
func (eer *externalEventReceiver) RolledBack(tag interface{}) {
	eer.manager.Queue() <- rolledBackEvent{}
}
