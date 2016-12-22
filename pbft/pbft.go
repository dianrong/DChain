// Copyright Dianrong.com Corp. 2016 All Rights Reserved.
//
// The Roc is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package pbft

import (
	"sync"
	"time"
	"fmt"
	"github.com/spf13/viper"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/common"
)

type Consenter interface {
	RecvMsg(*types.Message) error // Called serially with incoming messages from gRPC
}


type msgID struct { // our index through certStore
	v uint32
	n uint32
}

type msgCert struct {
	digest      common.Hash
	prePrepare  *types.PrePrepare
	sentPrepare bool
	prepare     []*types.Prepare
	sentCommit  bool
	commit      []*types.Commit
}

type pbftCore struct {
						       // internal data
	internalLock sync.Mutex
	executing    bool // signals that application is executing

	idleChan   chan struct{} // Used to detect idleness for testing
	injectChan chan func()   // Used as a hack to inject work onto the PBFT thread, to be removed eventually

	mux	*event.TypeMux

	activeView    bool              // view change happening
	byzantine     bool              // whether this node is intentionally acting as Byzantine; useful for debugging on the testnet
	f             uint32               // max. number of faults we can tolerate
	N             uint32               // max.number of validators in the network
	h             uint32            // low watermark
	id            uint32            // replica ID; PBFT `i`
	K             uint32            // checkpoint period
	logMultiplier uint32            // use this value to calculate log size : k*logMultiplier
	L             uint32            // log size
	lastExec      uint32            // last request we executed
	replicaCount  uint32               // number of replicas; PBFT `|R|`
	seqNo         uint32            // PBFT "n", strictly monotonic increasing sequence number
	view          uint32            // current view
	chkpts        map[uint32]string // state checkpoints; map lastExec to global hash


	skipInProgress    bool               // Set when we have detected a fall behind scenario until we pick a new starting point
	stateTransferring bool               // Set when state transfer is executing
	hChkpts           map[uint32]uint32  // highest checkpoint sequence number observed for each replica

	currentExec           *uint32                  // currently executing request
	timerActive           bool                     // is the timer running?
	//vcResendTimer         events.Timer             // timer triggering resend of a view change
	//newViewTimer          events.Timer             // timeout triggering a view change
	requestTimeout        time.Duration            // progress timeout for requests
	vcResendTimeout       time.Duration            // timeout before resending view change
	newViewTimeout        time.Duration            // progress timeout for new views
	newViewTimerReason    string                   // what triggered the timer
	lastNewViewTimeout    time.Duration            // last timeout we used during this view change
	broadcastTimeout      time.Duration            // progress timeout for broadcast
	outstandingReqBatches map[common.Hash]*types.RequestBatch // track whether we are waiting for request batches to execute

	//nullRequestTimer   events.Timer  // timeout triggering a null request
	nullRequestTimeout time.Duration // duration for this timeout
	viewChangePeriod   uint32        // period between automatic view changes
	viewChangeSeqNo    uint32        // next seqNo to perform view change

	missingReqBatches map[common.Hash]bool // for all the assigned, non-checkpointed request batches we might be missing during view-change

	// implementation of PBFT `in`
	reqBatchStore   map[common.Hash]*types.RequestBatch // track request batches
	certStore       map[msgID]*msgCert       // track quorum certificates for requests
	//checkpointStore map[Checkpoint]bool      // track checkpoints as set
	//viewChangeStore map[vcidx]*ViewChange    // track view-change messages
	//newViewStore    map[uint64]*NewView      // track last new-view we received or sent
}

func New(mux *event.TypeMux, peerId uint32, peerCount uint32) Consenter {
	return newObcBatch(mux, peerId, peerCount)
}

func newPbftCore(peerId uint32, peerCount uint32, mux *event.TypeMux) *pbftCore {
	var err error
	instance := &pbftCore{}
	instance.id = peerId
	instance.replicaCount = peerCount

	instance.mux = mux
	//instance.consumer = consumer
	//
	//instance.newViewTimer = etf.CreateTimer()
	//instance.vcResendTimer = etf.CreateTimer()
	//instance.nullRequestTimer = etf.CreateTimer()

	instance.N = uint32(viper.GetInt("consensus.N"))
	instance.f = uint32(viper.GetInt("consensus.f"))
	if instance.f*3+1 > instance.N {
		panic(fmt.Sprintf("need at least %d enough replicas to tolerate %d byzantine faults, but only %d replicas viperured", instance.f*3+1, instance.f, instance.N))
	}

	instance.K = uint32(viper.GetInt("consensus.K"))

	instance.logMultiplier = uint32(viper.GetInt("consensus.logmultiplier"))
	if instance.logMultiplier < 2 {
		panic("Log multiplier must be greater than or equal to 2")
	}
	instance.L = instance.logMultiplier * instance.K // log size
	instance.viewChangePeriod = uint32(viper.GetInt("consensus.viewchangeperiod"))

	instance.byzantine = viper.GetBool("consensus.byzantine")

	instance.requestTimeout, err = time.ParseDuration(viper.GetString("consensus.timeout.request"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse request timeout: %s", err))
	}
	instance.vcResendTimeout, err = time.ParseDuration(viper.GetString("consensus.timeout.resendviewchange"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse request timeout: %s", err))
	}
	instance.newViewTimeout, err = time.ParseDuration(viper.GetString("consensus.timeout.viewchange"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse new view timeout: %s", err))
	}
	instance.nullRequestTimeout, err = time.ParseDuration(viper.GetString("consensus.timeout.nullrequest"))
	if err != nil {
		instance.nullRequestTimeout = 0
	}
	instance.broadcastTimeout, err = time.ParseDuration(viper.GetString("consensus.timeout.broadcast"))
	if err != nil {
		panic(fmt.Errorf("Cannot parse new broadcast timeout: %s", err))
	}

	instance.activeView = true
	instance.replicaCount = instance.N

	//glog.Infof("PBFT type = %T", instance.consumer)
	glog.Infof("PBFT Max number of validating peers (N) = %v", instance.N)
	glog.Infof("PBFT Max number of failing peers (f) = %v", instance.f)
	glog.Infof("PBFT byzantine flag = %v", instance.byzantine)
	glog.Infof("PBFT request timeout = %v", instance.requestTimeout)
	glog.Infof("PBFT view change timeout = %v", instance.newViewTimeout)
	glog.Infof("PBFT Checkpoint period (K) = %v", instance.K)
	glog.Infof("PBFT broadcast timeout = %v", instance.broadcastTimeout)
	glog.Infof("PBFT Log multiplier = %v", instance.logMultiplier)
	glog.Infof("PBFT log size (L) = %v", instance.L)
	if instance.nullRequestTimeout > 0 {
		glog.Infof("PBFT null requests timeout = %v", instance.nullRequestTimeout)
	} else {
		glog.Infof("PBFT null requests disabled")
	}
	if instance.viewChangePeriod > 0 {
		glog.Infof("PBFT view change period = %v", instance.viewChangePeriod)
	} else {
		glog.Infof("PBFT automatic view change disabled")
	}

	// init the logs
	//instance.certStore = make(map[msgID]*msgCert)
	instance.reqBatchStore = make(map[common.Hash]*types.RequestBatch)
	//instance.checkpointStore = make(map[Checkpoint]bool)
	//instance.chkpts = make(map[uint64]string)
	//instance.viewChangeStore = make(map[vcidx]*ViewChange)
	//instance.pset = make(map[uint64]*ViewChange_PQ)
	//instance.qset = make(map[qidx]*ViewChange_PQ)
	//instance.newViewStore = make(map[uint64]*NewView)
	//
	//// initialize state transfer
	//instance.hChkpts = make(map[uint64]uint64)
	//
	//instance.chkpts[0] = "XXX GENESIS"
	//
	//instance.lastNewViewTimeout = instance.newViewTimeout
	instance.outstandingReqBatches = make(map[common.Hash]*types.RequestBatch)
	instance.missingReqBatches = make(map[common.Hash]bool)

	//instance.restoreState()
	//
	//instance.viewChangeSeqNo = ^uint64(0) // infinity
	//instance.updateViewChangeSeqNo()

	return instance
}

// allow the view-change protocol to kick-off when the timer expires
func (instance *pbftCore) ProcessEvent(e Event) Event {
	var err error
	logger.Debugf("Replica %d processing event", instance.id)
	switch et := e.(type) {
	case *types.RequestBatch:
		err = instance.recvRequestBatch(et)
	default:
		logger.Warningf("Replica %d received an unknown message type %T", instance.id, et)
	}

	if err != nil {
		logger.Warning(err.Error())
	}

	return nil
}

// Given a certain view n, what is the expected primary?
func (instance *pbftCore) primary(n uint32) uint32 {
	return n % uint32(instance.replicaCount)
}

// Is the sequence number between watermarks?
func (instance *pbftCore) inW(n uint32) bool {
	return n-instance.h > 0 && n-instance.h <= instance.L
}

// Is the view right? And is the sequence number between watermarks?
func (instance *pbftCore) inWV(v uint32, n uint32) bool {
	return instance.view == v && instance.inW(n)
}

// Given a digest/view/seq, is there an entry in the certLog?
// If so, return it. If not, create it.
func (instance *pbftCore) getCert(v uint32, n uint32) (cert *msgCert) {
	idx := msgID{v, n}
	cert, ok := instance.certStore[idx]
	if ok {
		return
	}

	cert = &msgCert{}
	instance.certStore[idx] = cert
	return
}

func (instance *pbftCore) recvRequestBatch(reqBatch *types.RequestBatch) error {
	digest, err := hash(reqBatch)
	if err != nil {
		return err
	}

	logger.Debugf("Replica %d received request batch %s", instance.id, digest)

	instance.reqBatchStore[digest] = reqBatch
	instance.outstandingReqBatches[digest] = reqBatch
	//instance.persistRequestBatch(digest)

	if instance.primary(instance.view) == instance.id && instance.activeView {
		instance.sendPrePrepare(reqBatch, digest)
	} else {
		logger.Debugf("Replica %d is backup, not sending pre-prepare for request batch %s", instance.id, digest)
	}
	return nil
}

func (instance *pbftCore) sendPrePrepare(reqBatch *types.RequestBatch, digest common.Hash) {
	logger.Debugf("Replica %d is primary, issuing pre-prepare for request batch %s", instance.id, digest)

	n := instance.seqNo + 1
	for _, cert := range instance.certStore { // check for other PRE-PREPARE for same digest, but different seqNo
		if p := cert.prePrepare; p != nil {
			if p.View == instance.view && p.SequenceNumber != n && p.BatchDigest == digest && digest.Str() != "" {
				logger.Infof("Other pre-prepare found with same digest but different seqNo: %d instead of %d", p.SequenceNumber, n)
				return
			}
		}
	}

	if !instance.inWV(instance.view, n) || n > instance.h+instance.L/2 {
		// We don't have the necessary stable certificates to advance our watermarks
		logger.Warningf("Primary %d not sending pre-prepare for batch %s - out of sequence numbers", instance.id, digest)
		return
	}

	if n > instance.viewChangeSeqNo {
		logger.Info("Primary %d about to switch to next primary, not sending pre-prepare with seqno=%d", instance.id, n)
		return
	}

	logger.Debugf("Primary %d broadcasting pre-prepare for view=%d/seqNo=%d and digest %s", instance.id, instance.view, n, digest)
	instance.seqNo = n
	preprep := &types.PrePrepare{
		View:           instance.view,
		SequenceNumber: n,
		BatchDigest:    digest,
		RequestBatch:   reqBatch,
		ReplicaId:      instance.id,
	}
	cert := instance.getCert(instance.view, n)
	cert.prePrepare = preprep
	cert.digest = digest
	//instance.persistQSet()

	// Broadcast the request to the network, in case we're in the wrong view
	instance.mux.Post(preprep)

	//instance.maybeSendCommit(digest, instance.view, n)
}
