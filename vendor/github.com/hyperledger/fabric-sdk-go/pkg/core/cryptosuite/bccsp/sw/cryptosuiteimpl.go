/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspSw "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/core")

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	// TODO: delete this check?
	if config.SecurityProvider() != "sw" {
		return nil, errors.Errorf("Unsupported BCCSP Provider: %s", config.SecurityProvider())
	}

	opts := getOptsByConfig(config)
	csp, err := getBCCSPFromOpts(opts)
	if err != nil {
		return nil, err
	}

	return wrapper.NewCryptoSuite(csp), nil
}

//GetSuiteWithDefaultEphemeral returns cryptosuite adaptor for bccsp with default ephemeral options (intended to aid testing)
func GetSuiteWithDefaultEphemeral() (core.CryptoSuite, error) {
	opts := getEphemeralOpts()
	csp, err := getBCCSPFromOpts(opts)
	if err != nil {
		return nil, err
	}

	return wrapper.NewCryptoSuite(csp), nil
}

func getBCCSPFromOpts(config *bccspSw.FactoryOpts) (bccsp.BCCSP, error) {
	f := &bccspSw.SWFactory{}
	csp, err := f.Get(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not initialize BCCSP %s", f.Name())
	}
	return csp, nil
}

// GetSuite returns a new instance of the software-based BCCSP
// set at the passed security level, hash family and KeyStore.
func GetSuite(securityLevel int, hashFamily string, keyStore bccsp.KeyStore) (core.CryptoSuite, error) {
	bccsp, err := sw.NewWithParams(securityLevel, hashFamily, keyStore)
	if err != nil {
		return nil, err
	}
	return wrapper.NewCryptoSuite(bccsp), nil
}

//GetOptsByConfig Returns Factory opts for given SDK config
func getOptsByConfig(c core.CryptoSuiteConfig) *bccspSw.FactoryOpts {
	opts := &bccspSw.SwOpts{
		HashFamily: c.SecurityAlgorithm(),
		SecLevel:   c.SecurityLevel(),
		// 从配置文件中读取加密插件类型
		Library: c.SecurityProviderLibPath(),
		FileKeystore: &bccspSw.FileKeystoreOpts{
			KeyStorePath: c.KeyStorePath(),
		},
	}
	logger.Debug("Initialized SW cryptosuite")

	return &bccspSw.FactoryOpts{SwOpts: opts}
}

func getEphemeralOpts() *bccspSw.FactoryOpts {
	opts := &bccspSw.SwOpts{
		HashFamily: "SHA2",
		SecLevel:   256,
	}
	logger.Debug("Initialized ephemeral SW cryptosuite with default opts")

	return &bccspSw.FactoryOpts{SwOpts: opts}
}
