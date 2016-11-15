package pbft

import (
	"time"
	"fmt"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/spf13/viper"
	"github.com/ethereum/go-ethereum/core/types"
)

type obcBatch struct {

	externalEventReceiver
	pbft        *pbftCore
	broadcaster *broadcaster

	batchSize        int
	batchTimerActive bool
	batchTimeout     time.Duration
}

type batchMessage struct {
	tx	*types.Transaction
}

// Event types

// batchMessageEvent is sent when a consensus message is received that is then to be sent to pbft
type batchMessageEvent batchMessage

func newObcBatch() *obcBatch {
	var err error

	op := &obcBatch{}
	op.pbft = newPbftCore()

	op.batchSize = viper.GetInt("general.batchsize")
	op.batchTimeout, err = time.ParseDuration(viper.GetString("general.timeout.batch"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse batch timeout: %s", err))
	}
	glog.Infof("PBFT Batch size = %d", op.batchSize)
	glog.Infof("PBFT Batch timeout = %v", op.batchTimeout)

	op.manager = NewManagerImpl() // TODO, this is hacky, eventually rip it out
	op.manager.SetReceiver(op)
	op.manager.Start()
	op.externalEventReceiver.manager = op.manager

	if op.batchTimeout >= op.pbft.requestTimeout {
		op.pbft.requestTimeout = 3 * op.batchTimeout / 2
		glog.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", op.pbft.requestTimeout)
	}

	if op.pbft.requestTimeout >= op.pbft.nullRequestTimeout && op.pbft.nullRequestTimeout != 0 {
		op.pbft.nullRequestTimeout = 3 * op.pbft.requestTimeout / 2
		glog.Warningf("Configured null request timeout must be greater than request timeout, setting to %v", op.pbft.nullRequestTimeout)
	}

	return op
}

// allow the primary to send a batch when the timer expires
func (op *obcBatch) ProcessEvent(event Event) Event {
	logger.Debugf("Replica %d batch main thread looping", op.pbft.id)
	switch et := event.(type) {
	case batchMessageEvent:
		msg := et
		return op.processMessage(msg.tx)
	default:
		return op.pbft.ProcessEvent(event)
	}

	return nil
}


func (op *obcBatch) processMessage(tx *types.Transaction) Event {
	fmt.Printf("transaction : data - %d;  value - %d", tx.Data(), tx.Value())
	return nil
}

