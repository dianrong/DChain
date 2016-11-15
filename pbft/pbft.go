package pbft

import (
	"sync"
	"time"
	"fmt"
	"github.com/spf13/viper"
	"github.com/ethereum/go-ethereum/logger/glog"
)

type pbftCore struct {
						       // internal data
	internalLock sync.Mutex
	executing    bool // signals that application is executing

	idleChan   chan struct{} // Used to detect idleness for testing
	injectChan chan func()   // Used as a hack to inject work onto the PBFT thread, to be removed eventually

						       // PBFT data
	activeView    bool              // view change happening
	byzantine     bool              // whether this node is intentionally acting as Byzantine; useful for debugging on the testnet
	f             int               // max. number of faults we can tolerate
	N             int               // max.number of validators in the network
	h             uint64            // low watermark
	id            uint64            // replica ID; PBFT `i`
	K             uint64            // checkpoint period
	logMultiplier uint64            // use this value to calculate log size : k*logMultiplier
	L             uint64            // log size
	lastExec      uint64            // last request we executed
	replicaCount  int               // number of replicas; PBFT `|R|`
	seqNo         uint64            // PBFT "n", strictly monotonic increasing sequence number
	view          uint64            // current view
	chkpts        map[uint64]string // state checkpoints; map lastExec to global hash


	skipInProgress    bool               // Set when we have detected a fall behind scenario until we pick a new starting point
	stateTransferring bool               // Set when state transfer is executing
	hChkpts           map[uint64]uint64  // highest checkpoint sequence number observed for each replica

	currentExec           *uint64                  // currently executing request
	timerActive           bool                     // is the timer running?
	//vcResendTimer         events.Timer             // timer triggering resend of a view change
	//newViewTimer          events.Timer             // timeout triggering a view change
	requestTimeout        time.Duration            // progress timeout for requests
	vcResendTimeout       time.Duration            // timeout before resending view change
	newViewTimeout        time.Duration            // progress timeout for new views
	newViewTimerReason    string                   // what triggered the timer
	lastNewViewTimeout    time.Duration            // last timeout we used during this view change
	broadcastTimeout      time.Duration            // progress timeout for broadcast
	//outstandingReqBatches map[string]*RequestBatch // track whether we are waiting for request batches to execute

	//nullRequestTimer   events.Timer  // timeout triggering a null request
	nullRequestTimeout time.Duration // duration for this timeout
	viewChangePeriod   uint64        // period between automatic view changes
	viewChangeSeqNo    uint64        // next seqNo to perform view change

	missingReqBatches map[string]bool // for all the assigned, non-checkpointed request batches we might be missing during view-change

	// implementation of PBFT `in`
	//reqBatchStore   map[string]*RequestBatch // track request batches
	//certStore       map[msgID]*msgCert       // track quorum certificates for requests
	//checkpointStore map[Checkpoint]bool      // track checkpoints as set
	//viewChangeStore map[vcidx]*ViewChange    // track view-change messages
	//newViewStore    map[uint64]*NewView      // track last new-view we received or sent
}

func newPbftCore(id uint64, /*consumer innerStack, etf events.TimerFactory*/) *pbftCore {
	var err error
	instance := &pbftCore{}
	instance.id = id
	//instance.consumer = consumer
	//
	//instance.newViewTimer = etf.CreateTimer()
	//instance.vcResendTimer = etf.CreateTimer()
	//instance.nullRequestTimer = etf.CreateTimer()

	instance.N = viper.GetInt("consensus.N")
	instance.f = viper.GetInt("consensus.f")
	if instance.f*3+1 > instance.N {
		panic(fmt.Sprintf("need at least %d enough replicas to tolerate %d byzantine faults, but only %d replicas viperured", instance.f*3+1, instance.f, instance.N))
	}

	instance.K = uint64(viper.GetInt("consensus.K"))

	instance.logMultiplier = uint64(viper.GetInt("consensus.logmultiplier"))
	if instance.logMultiplier < 2 {
		panic("Log multiplier must be greater than or equal to 2")
	}
	instance.L = instance.logMultiplier * instance.K // log size
	instance.viewChangePeriod = uint64(viper.GetInt("consensus.viewchangeperiod"))

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
	//instance.reqBatchStore = make(map[string]*RequestBatch)
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
	//instance.outstandingReqBatches = make(map[string]*RequestBatch)
	//instance.missingReqBatches = make(map[string]bool)
	//
	//instance.restoreState()
	//
	//instance.viewChangeSeqNo = ^uint64(0) // infinity
	//instance.updateViewChangeSeqNo()

	return instance
}



