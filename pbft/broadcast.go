// Copyright Dianrong.com Corp. 2016 All Rights Reserved.
//
// The DChain is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

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
