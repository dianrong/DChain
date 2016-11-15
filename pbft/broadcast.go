package pbft

import (
	"time"
	"sync"
)


type broadcaster struct {
	//comm communicator

	f                int
	broadcastTimeout time.Duration
	//msgChans         map[uint64]chan *sendRequest
	closed           sync.WaitGroup
	closedCh         chan struct{}
}
