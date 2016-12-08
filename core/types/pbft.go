package types

import (
	"time"
	"github.com/ethereum/go-ethereum/common"
)

type PrePrepare struct {
	View           uint32
	SequenceNumber uint32
	BatchDigest    common.Hash
	RequestBatch   *RequestBatch
	ReplicaId      uint32
}

type Prepare struct {
	View           uint32
	SequenceNumber uint32
	BatchDigest    common.Hash
	ReplicaId      uint32
}

type Commit struct {
	View           uint32
	SequenceNumber uint32
	BatchDigest    common.Hash
	ReplicaId      uint32
}


type BatchMessage struct {
	Msg	*Message
}

type RequestBatch struct {
	Batch []*Request
}

// batchMessageEvent is sent when a consensus message is received that is then to be sent to pbft
type BatchMessageEvent BatchMessage


type Message_Type int32

const (
	Message_UNDEFINED               Message_Type = 0
	Message_DISC_HELLO              Message_Type = 1
	Message_DISC_DISCONNECT         Message_Type = 2
	Message_DISC_GET_PEERS          Message_Type = 3
	Message_DISC_PEERS              Message_Type = 4
	Message_DISC_NEWMSG             Message_Type = 5
	Message_CHAIN_TRANSACTION       Message_Type = 6
	Message_SYNC_GET_BLOCKS         Message_Type = 11
	Message_SYNC_BLOCKS             Message_Type = 12
	Message_SYNC_BLOCK_ADDED        Message_Type = 13
	Message_SYNC_STATE_GET_SNAPSHOT Message_Type = 14
	Message_SYNC_STATE_SNAPSHOT     Message_Type = 15
	Message_SYNC_STATE_GET_DELTAS   Message_Type = 16
	Message_SYNC_STATE_DELTAS       Message_Type = 17
	Message_RESPONSE                Message_Type = 20
	Message_CONSENSUS               Message_Type = 21
)

type Message struct {
	Type 		Message_Type
	Tx   		*Transaction
	Prerepare	*PrePrepare
}

type Request struct {
	Timestamp 	time.Time
	Tx   		*Transaction
	ReplicaId 	uint32
}