package x509

import (
	"fmt"
	"strings"

	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/x509"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/x509/ccsgm"
)

// InitX509 load and save x509 plugin with name
var logger = flogging.MustGetLogger("X509-Plugin")

func InitX509(x509PluginType string) error {
	logger.Infof("InitX509Plugin: Plugin Name [%s]", x509PluginType)
	var plugin x509.X509
	switch strings.ToLower(x509PluginType) {
	// Default plugin
	case "":
		plugin = x509.NewStandardCert()
	case "std":
		plugin = x509.NewStandardCert()
	case "ccsgm":
		plugin = ccsgm.NewX509()
	case "sm":
		plugin = ccsgm.NewX509()
	default:
		return fmt.Errorf("unrecognized x509 plugin type: %s", x509PluginType)
	}

	return x509.AddPlugin(plugin)
}
