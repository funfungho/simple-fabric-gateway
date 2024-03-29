package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	configFilePath   = "./config.yaml"
	gmConfigFilePath = "./config-gm.yaml"

	systemChannelName   = "system-channel"
	genesisBlockProfile = "TwoOrgsOrdererGenesis"
	genesisBlock        = "/Users/slackbuffer/go/src/github.com/hyperledger/fabric/fabric-samples/test-network/system-genesis-block/genesis.block"
	ordererMSPDir       = "/Users/slackbuffer/go/src/github.com/hyperledger/fabric/fabric-samples/test-network/organizations/ordererOrganizations/example.com/msp"
	sdkOrgMSPDir        = "/Users/slackbuffer/go/src/github.com/hyperledger/fabric/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/msp"
	org2MSPDir          = "/Users/slackbuffer/go/src/github.com/hyperledger/fabric/fabric-samples/test-network/organizations/peerOrganizations/org2.example.com/msp"

	sdkOrg   = "Org1"
	sdkAdmin = "Admin"
)

var channelCreateTx = fmt.Sprintf("/Users/slackbuffer/go/src/github.com/hyperledger/fabric/fabric-samples/test-network/channel-artifacts/%s.tx", channelName)

func createGenesisBlock(w http.ResponseWriter, r *http.Request) {
	var err error
	if strings.HasPrefix(r.RequestURI, "/gm") {
		err = doCreateGenesisBlock(gmConfigFilePath)
	} else {
		err = doCreateGenesisBlock(configFilePath)
	}
	if err != nil {
		log.Println(err.Error())
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "create genesis block ok")
}

func doCreateGenesisBlock(sdkConfigFilePath string) error {
	sdk, err := fabsdk.New(config.FromFile(sdkConfigFilePath))
	if err != nil {
		return err
	}
	defer sdk.Close()
	clientContextProvider := sdk.Context(fabsdk.WithUser(sdkAdmin), fabsdk.WithOrg(sdkOrg))
	cc, err := clientContextProvider()
	if err != nil {
		return err
	}
	ho := cc.SigningManager().GetHashOpts()

	ordererOrg := &genesisconfig.Organization{
		Name:   "OrdererOrg",
		ID:     "OrdererMSP",
		MSPDir: ordererMSPDir,
	}
	sdkOrg := &genesisconfig.Organization{
		Name:   "Org1MSP",
		ID:     "Org1MSP",
		MSPDir: sdkOrgMSPDir,
	}
	org2 := &genesisconfig.Organization{
		Name:   "Org2MSP",
		ID:     "Org2MSP",
		MSPDir: org2MSPDir,
	}

	gc := &genesisconfig.GenesisConfig{
		ChainID:         systemChannelName,
		OrdererType:     "solo",
		Addresses:       []string{fmt.Sprintf("%s:%d", ordererEndpoint, ordererPort)},
		BatchTimeout:    2 * time.Second,
		MaxMessageCount: 10,
		// 2022-12-01 08:55:23.780 UTC [orderer.common.broadcast] ProcessMessage -> WARN 014 [channel: mychannel] Rejecting broadcast of config message from 172.26.0.1:35544 because of error: message payload is 22798 bytes and exceeds maximum allowed 4096 bytes
		AbsoluteMaxBytes:        103809024,
		PreferredMaxBytes:       524288,
		OrdererOrganizations:    []*genesisconfig.Organization{ordererOrg},
		ConsortiumOrganizations: []*genesisconfig.Organization{sdkOrg, org2},
		AdminsPolicy:            genesisconfig.PolicyAnyAdmins,
		WritersPolicy:           genesisconfig.PolicyAllWriters,
		ReadersPolicy:           genesisconfig.PolicyAllReaders,
	}
	profile := genesisconfig.NewGenesisProfile(gc)
	gbbs, err := resource.CreateGenesisBlockForOrdererWithHashOpts(profile, systemChannelName, cc.CryptoSuite(), ho)
	if err != nil {
		return err
	}

	err = writeFile(genesisBlock, gbbs, 0640)
	if err != nil {
		return fmt.Errorf("error writing genesis block: %s", err)
	}

	return nil
}

func createChannelCreateTx(w http.ResponseWriter, r *http.Request) {
	var err error
	if strings.HasPrefix(r.RequestURI, "/gm") {
		err = doCreateChannelCreateTx(gmConfigFilePath)
	} else {
		err = doCreateChannelCreateTx(configFilePath)
	}
	if err != nil {
		log.Println(err.Error())
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "create channel create tx ok")
}

func doCreateChannelCreateTx(sdkConfigFilePath string) error {
	// sdk, err := fabsdk.New(config.FromFile(sdkConfigFilePath))
	// if err != nil {
	// 	return err
	// }
	// defer sdk.Close()
	// clientContextProvider := sdk.Context(fabsdk.WithUser(sdkAdmin), fabsdk.WithOrg(sdkOrg))

	// cc, err := clientContextProvider()
	// if err != nil {
	// 	return err
	// }
	// ho := cc.SigningManager().GetHashOpts()

	sdkOrg := &genesisconfig.Organization{
		Name:    "Org1MSP",
		ID:      "Org1MSP",
		MSPDir:  sdkOrgMSPDir,
		MSPType: "bccsp",
		Policies: map[string]*genesisconfig.Policy{
			"Admins": {
				Type: "Signature",
				Rule: "OR('Org1MSP.admin')",
			},
			"Readers": {
				Type: "Signature",
				Rule: "OR('Org1MSP.admin', 'Org1MSP.peer', 'Org1MSP.client')",
			},
			"Writers": {
				Type: "Signature",
				Rule: "OR('Org1MSP.admin', 'Org1MSP.client')",
			},
			"Endorsement": {
				Type: "Signature",
				Rule: "OR('Org1MSP.peer')",
			},
		},
	}
	org2 := &genesisconfig.Organization{
		Name:    "Org2MSP",
		ID:      "Org2MSP",
		MSPDir:  org2MSPDir,
		MSPType: "bccsp",
		Policies: map[string]*genesisconfig.Policy{
			"Admins": {
				Type: "Signature",
				Rule: "OR('Org2MSP.admin')",
			},
			"Readers": {
				Type: "Signature",
				Rule: "OR('Org2MSP.admin', 'Org2MSP.peer', 'Org2MSP.client')",
			},
			"Writers": {
				Type: "Signature",
				Rule: "OR('Org2MSP.admin', 'Org2MSP.client')",
			},
			"Endorsement": {
				Type: "Signature",
				Rule: "OR('Org2MSP.peer')",
			},
		},
	}

	gp := &genesisconfig.Profile{
		Consortium: "SampleConsortium",
		Policies: map[string]*genesisconfig.Policy{
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "ANY Admins",
			},
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Endorsement": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
		},
		Application: &genesisconfig.Application{
			ACLs: map[string]string{
				"_lifecycle/CommitChaincodeDefinition": "/Channel/Application/Writers",
				"_lifecycle/QueryChaincodeDefinition":  "/Channel/Application/Readers",
				"_lifecycle/QueryNamespaceDefinitions": "/Channel/Application/Readers",
				"lscc/ChaincodeExists":                 "/Channel/Application/Readers",
				"lscc/GetDeploymentSpec":               "/Channel/Application/Readers",
				"lscc/GetChaincodeData":                "/Channel/Application/Readers",
				"lscc/GetInstantiatedChaincodes":       "/Channel/Application/Readers",
				"qscc/GetChainInfo":                    "/Channel/Application/Readers",
				"qscc/GetBlockByNumber":                "/Channel/Application/Readers",
				"qscc/GetBlockByHash":                  "/Channel/Application/Readers",
				"qscc/GetTransactionByID":              "/Channel/Application/Readers",
				"qscc/GetBlockByTxID":                  "/Channel/Application/Readers",
				"cscc/GetConfigBlock":                  "/Channel/Application/Readers",
				"cscc/GetConfigTree":                   "/Channel/Application/Readers",
				"cscc/SimulateConfigTreeUpdate":        "/Channel/Application/Readers",
				"peer/Propose":                         "/Channel/Application/Writers",
				"peer/ChaincodeToChaincode":            "/Channel/Application/Readers",
				"event/Block":                          "/Channel/Application/Readers",
				"event/FilteredBlock":                  "/Channel/Application/Readers",
			},
			Organizations: []*genesisconfig.Organization{sdkOrg, org2},
			Policies: map[string]*genesisconfig.Policy{
				"LifecycleEndorsement": {
					Type: "Signature",
					Rule: "OR('Org1MSP.peer')",
				},
				"Endorsement": {
					Type: "Signature",
					Rule: "OR('Org1MSP.peer')",
				},
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "ANY Admins",
				},
			},
			Capabilities: map[string]bool{
				"V2_0": true,
			},
		},
	}
	cctBytes, err := resource.CreateChannelCreateTx(gp, nil, channelName)
	if err != nil {
		return err
	}
	err = writeFile(channelCreateTx, cctBytes, 0640)
	if err != nil {
		return fmt.Errorf("error writing channel create transaction error: %s", err)
	}

	return nil
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(filename)
	exists, err := dirExists(dirPath)
	if err != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(dirPath, 0750)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filename, data, perm)
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// func getFabricCryptoConfig(w http.ResponseWriter, r *http.Request) {
// 	sdk, err := fabsdk.New(config.FromFile(configFilePath))
// 	if err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}

// 	cliContext := sdk.Context(fabsdk.WithUser(sdkAdmin), fabsdk.WithOrg(sdkOrg))
// 	rc, err := resmgmt.New(cliContext)
// 	if err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}

// 	cfgBlk, err := rc.QueryConfigBlockFromOrderer(channelName, resmgmt.WithOrdererEndpoint("orderer.example.com"))
// 	if err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}
// 	cfgBlkPayload := cfgBlk.Data.Data[0]
// 	envelope := common.Envelope{}
// 	if err := proto.Unmarshal(cfgBlkPayload, &envelope); err != nil {
// 		io.WriteString(w, err.Error())
// 	}
// 	payload := &common.Payload{}
// 	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}
// 	cfgEnv := &common.ConfigEnvelope{}
// 	if err := proto.Unmarshal(payload.Data, cfgEnv); err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}
// 	origCfg := cfgEnv.Config
// 	newCfg := proto.Clone(origCfg).(*common.Config)

// 	newMSPCfgBytes := newCfg.ChannelGroup.Groups[channelconfig.ApplicationGroupKey].Groups["Org1MSP"].Values[channelconfig.MSPKey].Value
// 	var newMSPCfg fabmsp.MSPConfig
// 	if err := proto.Unmarshal(newMSPCfgBytes, &newMSPCfg); err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}
// 	origFabMSPCfg := newMSPCfg.Config
// 	var fabMSPCfg fabmsp.FabricMSPConfig
// 	if err := proto.Unmarshal(origFabMSPCfg, &fabMSPCfg); err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}
// 	fcc, err := json.Marshal(fabMSPCfg.CryptoConfig)
// 	if err != nil {
// 		io.WriteString(w, err.Error())
// 		return
// 	}
// 	io.WriteString(w, string(fcc))
// }
