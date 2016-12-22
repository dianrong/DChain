// Copyright Dianrong.com Corp. 2016 All Rights Reserved.
//
// The Roc is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

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
