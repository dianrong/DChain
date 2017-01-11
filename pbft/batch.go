// Copyright Dianrong.com Corp. 2016 All Rights Reserved.
//
// The DChain is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package pbft

import (
	"time"
	"fmt"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/spf13/viper"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/core"
)

type obcBatch struct {

	externalEventReceiver

	mux         *event.TypeMux
	pbft        *pbftCore
	broadcaster *broadcaster

	batchSize        int
	batchTimerActive bool
	batchTimeout     time.Duration

	batchStore       []*types.Request
}


func newObcBatch(mux *event.TypeMux, peerId uint32, peerCount uint32) *obcBatch {
	var err error

	op := &obcBatch{}
	op.mux = mux
	op.pbft = newPbftCore(peerId, peerCount, mux)

	op.batchSize = viper.GetInt("consensus.batchsize")
	op.batchTimeout, err = time.ParseDuration(viper.GetString("consensus.timeout.batch"))
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
	case types.BatchMessageEvent:
		msg := et
		return op.processMessage(msg.Msg)
	default:
		return op.pbft.ProcessEvent(event)
	}

	return nil
}


func (op *obcBatch) processMessage(msg	*types.Message) Event {
	if msg.Type == types.Message_CHAIN_TRANSACTION {
		logger.Infof("transaction : data - %d;  value - %d", msg.Tx.Data(), msg.Tx.Value())

		// Broadcast the request to the network, in case we're in the wrong view
		op.mux.Post(core.TxPbftEvent{Tx: msg.Tx})
		return op.submitToLeader(msg.Tx)
	}

	if msg.Type != types.Message_CONSENSUS {
		logger.Errorf("Unexpected message type: %s", msg.Type)
		return nil
	}

	// TODO recive the Message_CONSENSUS from handler.go:
	if msg.Tx != nil {
		return op.submitToLeader(msg.Tx)
	} else if msg.Prerepare != nil {
		logger.Infof("recive the Message_CONSENSUS from handler.go : %s", msg.Prerepare.BatchDigest)
		logger.Infof("recive the Prerepare from handler.go : ", msg.Prerepare.BatchDigest)
	} else {
		logger.Infof("recive empty msg ")
	}

	return nil
}

func (op *obcBatch) submitToLeader(tx *types.Transaction) Event {

	logger.Infof("view id : %d; node id : %d", op.pbft.view, op.pbft.id)
	if op.pbft.primary(op.pbft.view) == op.pbft.id && op.pbft.activeView {
		logger.Infof("find primary node")
		return op.leaderProcReq(op.txToReq(tx))
	}

	return nil
}


func (op *obcBatch) txToReq(tx *types.Transaction) *types.Request {
	now := time.Now()
	req := &types.Request{
		Timestamp: now,
		Tx:   tx,
		ReplicaId: op.pbft.id,
	}

	return req
}

// =============================================================================
// functions specific to batch mode
// =============================================================================

func (op *obcBatch) leaderProcReq(req *types.Request) Event {
	// XXX check req sig

	logger.Debugf("Batch primary %d queueing new request", op.pbft.id)
	op.batchStore = append(op.batchStore, req)

	if len(op.batchStore) >= op.batchSize {
		return op.sendBatch()
	}

	return nil
}


func (op *obcBatch) sendBatch() Event {

	if len(op.batchStore) == 0 {
		logger.Error("Told to send an empty batch store for ordering, ignoring")
		return nil
	}

	reqBatch := &types.RequestBatch{Batch: op.batchStore}
	op.batchStore = nil
	logger.Infof("Creating batch with %d requests", len(reqBatch.Batch))
	return reqBatch
}

