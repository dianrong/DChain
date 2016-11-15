package pbft

import (
	"time"
	"fmt"
	"gopkg.in/karalabe/cookiejar.v2/collections/stack"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/spf13/viper"
)

type obcBatch struct {
	pbft  *pbftCore

	externalEventReceiver
	pbft        *pbftCore
	broadcaster *broadcaster

	batchSize        int
	batchTimerActive bool
	batchTimeout     time.Duration
}

func newObcBatch(id uint64) *obcBatch {
	var err error

	op := &obcBatch{}
	op.pbft = newPbftCore(id)

	op.batchSize = viper.GetInt("general.batchsize")
	op.batchTimeout, err = time.ParseDuration(viper.GetString("general.timeout.batch"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse batch timeout: %s", err))
	}
	glog.Infof("PBFT Batch size = %d", op.batchSize)
	glog.Infof("PBFT Batch timeout = %v", op.batchTimeout)

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
