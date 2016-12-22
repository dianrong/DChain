// Copyright Dianrong.com Corp. 2016 All Rights Reserved.
//
// The Roc is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package ca

import (
	"github.com/op/go-logging"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"github.com/spf13/viper"
)

var (
	log = logging.MustGetLogger("crypto")
)

// Init initializes the crypto layer. It load from viper the security level
// and the logging setting.
func Init() (err error) {
	// Init security level
	securityLevel := 256
	if viper.IsSet("security.level") {
		ovveride := viper.GetInt("security.level")
		if ovveride != 0 {
			securityLevel = ovveride
		}
	}

	hashAlgorithm := "SHA3"
	if viper.IsSet("security.hashAlgorithm") {
		ovveride := viper.GetString("security.hashAlgorithm")
		if ovveride != "" {
			hashAlgorithm = ovveride
		}
	}

	log.Debugf("Working at security level [%d]", securityLevel)

	if err = primitives.InitSecurityLevel(hashAlgorithm, securityLevel); err != nil {
		log.Errorf("Failed setting security level: [%s]", err)

		return
	}

	return
}
