package pbft

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"crypto/md5"
)

func hash(msg interface{}) (common.Hash, error) {
	data, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return common.Hash{}, err
	}

	sum := md5.Sum(data)
	hash := common.BytesToHash([]byte(sum[:]))

	return hash, nil
}
