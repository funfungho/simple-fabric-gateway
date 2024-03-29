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
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/crypto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/gm/ccsgm"
)

func NewFileBasedKeyStore(pwd []byte, path, implType string, readOnly bool) (bccsp.KeyStore, error) {
	ks := &fileBasedKeyStore{implType: implType}
	return ks, ks.Init(pwd, path, readOnly)
}

// fileBasedKeyStore is a folder-based KeyStore.
// Each key is stored in a separated file whose name contains the key's SKI
// and flags to identity the key's type. All the keys are stored in
// a folder whose path is provided at initialization time.
// The KeyStore can be initialized with a password, this password
// is used to encrypt and decrypt the files storing the keys.
// A KeyStore can be read only to avoid the overwriting of keys.
type fileBasedKeyStore struct {
	path     string
	implType string // ccsgm

	readOnly bool
	isOpen   bool

	pwd []byte

	// Sync
	m sync.Mutex
}

// Init initializes this KeyStore with a password, a path to a folder
// where the keys are stored and a read only flag.
// Each key is stored in a separated file whose name contains the key's SKI
// and flags to identity the key's type.
// If the KeyStore is initialized with a password, this password
// is used to encrypt and decrypt the files storing the keys.
// The pwd can be nil for non-encrypted KeyStores. If an encrypted
// key-store is initialized without a password, then retrieving keys from the
// KeyStore will fail.
// A KeyStore can be read only to avoid the overwriting of keys.
func (ks *fileBasedKeyStore) Init(pwd []byte, path string, readOnly bool) error {
	// Validate inputs
	// pwd can be nil

	if len(path) == 0 {
		return errors.New("An invalid KeyStore path provided. Path cannot be an empty string.")
	}

	ks.m.Lock()
	defer ks.m.Unlock()

	if ks.isOpen {
		return errors.New("KeyStore already initilized.")
	}

	ks.path = path
	ks.pwd = utils.Clone(pwd)

	err := ks.createKeyStoreIfNotExists()
	if err != nil {
		return err
	}

	err = ks.openKeyStore()
	if err != nil {
		return err
	}

	ks.readOnly = readOnly

	return nil
}

// ReadOnly returns true if this KeyStore is read only, false otherwise.
// If ReadOnly is true then StoreKey will fail.
func (ks *fileBasedKeyStore) ReadOnly() bool {
	return ks.readOnly
}

// GetKey returns a key object whose SKI is the one passed.
func (ks *fileBasedKeyStore) GetKey(ski []byte) (bccsp.Key, error) {
	// Validate arguments
	if len(ski) == 0 {
		return nil, errors.New("Invalid SKI. Cannot be of zero length.")
	}

	suffix := ks.getSuffix(hex.EncodeToString(ski))

	switch suffix {
	case "key":
		// Load the key
		path := ks.getPathForAlias(hex.EncodeToString(ski), "key")
		var key []byte
		var err error
		switch ks.implType {
		case "", "ccsgm":
			key, err = ccsgm.LoadKeyFromPem(path, nil)
		default:
			key, err = nil, fmt.Errorf("unsupported implType: [%s]", ks.implType)
		}
		if err != nil || key == nil {
			return nil, fmt.Errorf("Failed loading key [%x] [%s]", ski, err)
		}
		return &sm4PrivateKey{key, ski}, nil
	case "sk":
		// Load the private key

		path := ks.getPathForAlias(hex.EncodeToString(ski), "sk")
		var key *crypto.PrivateKey
		var err error
		switch ks.implType {
		case "", "ccsgm":
			key, err = ccsgm.LoadPrivateKeyFromPem(path, nil)
		default:
			key, err = nil, fmt.Errorf("unsupported implType: [%s]", ks.implType)
		}
		if err != nil || key == nil {
			return nil, fmt.Errorf("Failed loading secret key [%x] [%s]", ski, err)
		}
		return &sm2PrivateKey{key, ski}, nil

	case "pk":
		// Load the public key
		path := ks.getPathForAlias(hex.EncodeToString(ski), "pk")
		var key *crypto.PublicKey
		var err error
		switch ks.implType {
		case "", "ccsgm":
			key, err = ccsgm.LoadPublicKeyFromPem(path, nil)
		default:
			key, err = nil, fmt.Errorf("unsupported implType: [%s]", ks.implType)
		}
		if err != nil || key == nil {
			return nil, fmt.Errorf("Failed loading public key [%x] [%s]", ski, err)
		}
		return &sm2PublicKey{key, ski}, nil

	default:
		return ks.searchKeystoreForSKI(ski)
	}
}

// StoreKey stores the key k in this KeyStore.
// If this KeyStore is read only then the method will fail.
func (ks *fileBasedKeyStore) StoreKey(k bccsp.Key) (err error) {
	if ks.readOnly {
		return errors.New("Read only KeyStore.")
	}

	if k == nil {
		return errors.New("Invalid key. It must be different from nil.")
	}
	switch k.(type) {
	case *sm2PrivateKey:
		kk := k.(*sm2PrivateKey)
		if kk.privKey == nil {
			return errors.New("Invalid key. It's privkey must be different from nil")
		}
		path := ks.getPathForAlias(hex.EncodeToString(k.SKI()), "sk")
		switch ks.implType {
		case "", "ccsgm":
			_, err = ccsgm.SavePrivateKeytoPem(path, kk.privKey, nil)
		default:
			err = fmt.Errorf("unsupported implType: [%s]", ks.implType)
		}
		if err != nil {
			return fmt.Errorf("Failed storing sm2 private key [%s]", err)
		}

	case *sm2PublicKey:
		kk := k.(*sm2PublicKey)
		if kk.pubKey == nil {
			return errors.New("Invalid key. It's pubKey must be different from nil")
		}
		path := ks.getPathForAlias(hex.EncodeToString(k.SKI()), "pk")
		switch ks.implType {
		case "", "ccsgm":
			_, err = ccsgm.SavePublicKeytoPem(path, kk.pubKey, nil)
		default:
			err = fmt.Errorf("unsupported implType: [%s]", ks.implType)
		}
		if err != nil {
			return fmt.Errorf("Failed storing sm2 public key [%s]", err)
		}
	case *xinAnPrivateKey:
	case *sm4PrivateKey:
		kk := k.(*sm4PrivateKey)
		if kk.key == nil {
			return errors.New("Invalid key. It's key must be different from nil")
		}
		path := ks.getPathForAlias(hex.EncodeToString(k.SKI()), "key")
		switch ks.implType {
		case "", "ccsgm":
			_, err = ccsgm.SaveKeyToPem(path, kk.key, nil)
		default:
			err = fmt.Errorf("unsupported implType: [%s]", ks.implType)
		}
		if err != nil {
			return fmt.Errorf("Failed storing sm4 private key [%s]", err)
		}
	default:
		return fmt.Errorf("Key type not reconigned [%s]", k)
	}

	return
}

func (ks *fileBasedKeyStore) searchKeystoreForSKI(ski []byte) (k bccsp.Key, err error) {

	files, _ := ioutil.ReadDir(ks.path)
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		if f.Size() > (1 << 16) { //64k, somewhat arbitrary limit, considering even large RSA keys
			continue
		}

		var sk *crypto.PrivateKey
		var err error
		switch ks.implType {
		case "", "ccsgm":
			sk, err = ccsgm.LoadPrivateKeyFromPem(filepath.Join(ks.path, f.Name()), nil)
		default:
			sk, err = nil, fmt.Errorf("unsupported implType: [%s]", ks.implType)
		}
		if err != nil {
			continue
		}
		k = &sm2PrivateKey{sk, ski}
		if !bytes.Equal(k.SKI(), ski) {
			continue
		}

		return k, nil
	}
	return nil, fmt.Errorf("Key with SKI %s not found in %s", hex.EncodeToString(ski), ks.path)
}

func (ks *fileBasedKeyStore) getSuffix(alias string) string {
	files, _ := ioutil.ReadDir(ks.path)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), alias) {
			if strings.HasSuffix(f.Name(), "sk") {
				return "sk"
			}
			if strings.HasSuffix(f.Name(), "pk") {
				return "pk"
			}
			if strings.HasSuffix(f.Name(), "key") {
				return "key"
			}
			break
		}
	}
	return ""
}

func (ks *fileBasedKeyStore) createKeyStoreIfNotExists() error {
	// Check keystore directory
	ksPath := ks.path
	missing, _ := utils.DirMissingOrEmpty(ksPath)

	if missing {
		err := os.MkdirAll(ks.path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ks *fileBasedKeyStore) openKeyStore() error {
	if ks.isOpen {
		return nil
	}
	ks.isOpen = true

	return nil
}

func (ks *fileBasedKeyStore) getPathForAlias(alias, suffix string) string {
	return filepath.Join(ks.path, alias+"_"+suffix)
}
