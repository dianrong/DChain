// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// +build !opencl

package eth

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/caserver/ca"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/spf13/viper"
)

const disabledInfo = "Set GO_OPENCL and re-build to enable."

func (s *Ethereum) StartMining(threads int, gpus string) error {
	if s.nodetype != ca.Validator && s.nodetype != ca.Admin {
		err := fmt.Errorf("Cannot start mining without permission")
		glog.V(logger.Error).Infoln(err)
		return err
	}

	if viper.GetString("consensus.algorithm") != "POW" {
		err := fmt.Errorf("Cannot start mining without POW mode")
		glog.V(logger.Error).Infoln(err)
		return err
	}

	eb, err := s.Etherbase()
	if err != nil {
		err = fmt.Errorf("Cannot start mining without etherbase address: %v", err)
		glog.V(logger.Error).Infoln(err)
		return err
	}

	if gpus != "" {
		return errors.New("GPU mining disabled. " + disabledInfo)
	}

	// CPU mining
	go s.miner.Start(eb, threads)
	return nil
}

func GPUBench(gpuid uint64) {
	fmt.Println("GPU mining disabled. " + disabledInfo)
}

func PrintOpenCLDevices() {
	fmt.Println("OpenCL disabled. " + disabledInfo)
}
