/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package gm

import (
	"errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/gm"
)

type sm4PrivateKey struct {
	key []byte
	ski []byte
}

func (k *sm4PrivateKey) Bytes() (raw []byte, err error) {
	return k.key, nil
}

func (k *sm4PrivateKey) SKI() (ski []byte) {
	// Hash it
	ski, err := gm.NewSm3().Hash(k.key)
	if err != nil {
		ski = nil
	}
	return ski
}

func (k *sm4PrivateKey) Symmetric() bool {
	return true
}

func (k *sm4PrivateKey) Private() bool {
	return true
}

func (k *sm4PrivateKey) PublicKey() (bccsp.Key, error) {
	return nil, errors.New("Cannot call this method on a symmetric key.")
}
