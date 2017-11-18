package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type ServiceConfig struct {
	Version  string                 `yaml:"version,flow"`
	Network  map[string]interface{} `yaml:"networks,omitempty"`
	Services map[string]interface{} `yaml:"services"`
}

type Container struct {
	Image         string            `yaml:"image,omitempty"`
	ContainerName string            `yaml:"container_name,omitempty"`
	TTY           bool              `yaml:"tty,omitempty"`
	Extends       map[string]string `yaml:"extends,omitempty"`
	Environment   []string          `yaml:"environment,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	Command       string            `yaml:"command,omitempty"`

	Volumns  []string `yaml:"volumes,omitempty"`
	Ports    []string `yaml:"ports,omitempty"`
	Depends  []string `yaml:"depends_on,omitempty"`
	Networks []string `yaml:"networks,omitempty"`
}

func GenerateDockerFiles(networkConfigByte []byte, dirpath string, addCA bool) bool {
	networkConfig := make(map[string]interface{})
	json.Unmarshal(networkConfigByte, &networkConfig)
	orgConfigs, _ := networkConfig["orgs"].([]interface{})
	couchCount := 0
	portMap := generatePorts([]int{7051, 7053})
	couchPortMap := generatePorts([]int{5984})
	caPortMap := generatePorts([]int{7054})
	var serviceConf ServiceConfig
	serviceConf.Version = "2"
	netWrk := make(map[string]interface{})

	netWrk["fabricnetwork"] = make(map[string]string)
	serviceConf.Network = netWrk
	containers := make(map[string]interface{})
	cliDependency := make([]string, 0)
	//Add the orderer
	ordConfigInt, _ := networkConfig["orderers"].(interface{})
	ordConfig, _ := ordConfigInt.(map[string]interface{})
	orderContainer := BuildOrderer(dirpath, getString(ordConfig["ordererHostname"]), getString(ordConfig["domain"]))
	containers[orderContainer.ContainerName] = orderContainer
	cliDependency = append(cliDependency, orderContainer.ContainerName)
	//Adding the peers and couchdb
	for index, org := range orgConfigs {
		orgConfig, _ := org.(map[string]interface{})
		fmt.Printf("Processing org %d \n", index)
		peerCountFlt, _ := orgConfig["peerCount"].(float64)
		peerCount := int(peerCountFlt)
		fmt.Printf(" Peer count is %d \n ", peerCount)
		if addCA == true {
			caContainer := BuildCAImage(dirpath, getString(orgConfig["domain"]), getString(orgConfig["name"]), caPortMap[index])
			containers[caContainer.ContainerName] = caContainer
		}
		for peerIndex := 0; peerIndex < peerCount; peerIndex++ {
			peerID := fmt.Sprintf("peer%d", peerIndex)
			couchID := fmt.Sprintf("couch%d", couchCount)
			ports := portMap[couchCount]
			couchContainer := BuildCouchDB(couchID, couchPortMap[couchCount])
			containerImage := BuildPeerImage(dirpath, peerID, getString(orgConfig["domain"]), getString(orgConfig["mspID"]), couchID, orderContainer.ContainerName, ports)
			containers[couchContainer.ContainerName] = couchContainer
			containers[containerImage.ContainerName] = containerImage
			cliDependency = append(cliDependency, containerImage.ContainerName)
			couchCount++

		}
	}
	cli := BuildCLI(dirpath, cliDependency)
	containers[cli.ContainerName] = cli
	serviceConf.Services = containers
	serviceBytes, _ := yaml.Marshal(serviceConf)
	if addCA == true {
		ioutil.WriteFile(dirpath+"/docker-compose-template.yaml", serviceBytes, 0666)
	} else {
		ioutil.WriteFile(dirpath+"/docker-compose.yaml", serviceBytes, 0666)
	}
	//generate the base.yaml
	outBytes, _ := yaml.Marshal(BuildBaseImage(addCA))
	ioutil.WriteFile(dirpath+"/base.yaml", outBytes, 0666)

	return true

}
func BuildCLI(dirPath string, otherConatiners []string) Container {
	var cli Container
	cli.ContainerName = "cli"
	cli.Image = "hyperledger/fabric-tools:${IMAGE_TAG}"
	cli.TTY = true
	cli.WorkingDir = "/opt/ws"
	vols := make([]string, 0)
	vols = append(vols, "/var/run/:/host/var/run/")
	vols = append(vols, "./:/opt/ws")
	vols = append(vols, "./chaincode/github.com:/opt/gopath/src/github.com")

	cliEnvironment := make([]string, 0)
	cliEnvironment = append(cliEnvironment, "CORE_PEER_TLS_ENABLED=true")
	cliEnvironment = append(cliEnvironment, "GOPATH=/opt/gopath")
	cliEnvironment = append(cliEnvironment, "CORE_LOGGING_LEVEL=DEBUG")
	cliEnvironment = append(cliEnvironment, "CORE_PEER_ID=cli")

	cli.Environment = cliEnvironment
	cli.Volumns = vols
	cli.Depends = otherConatiners
	var networks = make([]string, 0)
	networks = append(networks, "fabricnetwork")
	cli.Networks = networks
	return cli

}
func BuildOrderer(cryptoBasePath, ordererName, domainName string) Container {

	extnds := make(map[string]string)
	extnds["file"] = "base.yaml"
	extnds["service"] = "orderer"
	ordFQDN := ordererName + "." + domainName
	vols := make([]string, 0)
	vols = append(vols, cryptoBasePath+"/genesis.block:/var/hyperledger/orderer/genesis.block")
	vols = append(vols, cryptoBasePath+"/crypto-config/ordererOrganizations/"+domainName+"/orderers/"+ordFQDN+"/msp:/var/hyperledger/orderer/msp")
	vols = append(vols, cryptoBasePath+"/crypto-config/ordererOrganizations/"+domainName+"/orderers/"+ordFQDN+"/tls/:/var/hyperledger/orderer/tls")
	var networks = make([]string, 0)
	networks = append(networks, "fabricnetwork")
	var ports = make([]string, 0)
	ports = append(ports, "7050:7050")
	var orderer Container
	orderer.ContainerName = ordFQDN
	orderer.Extends = extnds
	orderer.Volumns = vols
	orderer.Ports = ports
	orderer.Networks = networks
	return orderer
}
func BuildPeerImage(cryptoBasePath, peerId, domainName, mspID, couchID, ordererFQDN string, ports []string) Container {

	extnds := make(map[string]string)
	extnds["file"] = "base.yaml"
	extnds["service"] = "peer"
	peerFQDN := peerId + "." + domainName

	peerEnvironment := make([]string, 0)
	peerEnvironment = append(peerEnvironment, "CORE_PEER_ID="+peerFQDN)
	peerEnvironment = append(peerEnvironment, "CORE_PEER_ADDRESS="+peerFQDN+":7051")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_CHAINCODELISTENADDRESS="+peerFQDN+":7052")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_GOSSIP_EXTERNALENDPOINT="+peerFQDN+":7051")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_LOCALMSPID="+mspID)
	peerEnvironment = append(peerEnvironment, "CORE_LEDGER_STATE_STATEDATABASE=CouchDB")
	peerEnvironment = append(peerEnvironment, "CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS="+couchID+":5984")
	if peerId == "peer0" {
		peerEnvironment = append(peerEnvironment, "CORE_PEER_GOSSIP_BOOTSTRAP="+peerFQDN+":7051")
	} else {
		peerEnvironment = append(peerEnvironment, "CORE_PEER_GOSSIP_BOOTSTRAP=peer0."+domainName+":7051")
	}
	vols := make([]string, 0)
	vols = append(vols, "/var/run/:/host/var/run/")
	vols = append(vols, cryptoBasePath+"/crypto-config/peerOrganizations/"+domainName+"/peers/"+peerFQDN+"/msp:/etc/hyperledger/fabric/msp")
	vols = append(vols, cryptoBasePath+"/crypto-config/peerOrganizations/"+domainName+"/peers/"+peerFQDN+"/tls:/etc/hyperledger/fabric/tls")
	var depends = make([]string, 0)
	depends = append(depends, couchID)
	depends = append(depends, ordererFQDN)
	var networks = make([]string, 0)
	networks = append(networks, "fabricnetwork")

	var container Container
	container.ContainerName = peerFQDN
	container.Environment = peerEnvironment
	container.Volumns = vols
	container.Depends = depends
	container.Networks = networks
	container.Ports = ports
	container.Extends = extnds
	return container
}
func BuildCAImage(cryptoBasePath, domainName, orgname string, ports []string) Container {

	extnds := make(map[string]string)
	extnds["file"] = "base.yaml"
	extnds["service"] = "ca"
	peerFQDN := "ca." + domainName

	peerEnvironment := make([]string, 0)
	peerEnvironment = append(peerEnvironment, "FABRIC_CA_SERVER_CA_CERTFILE=/etc/hyperledger/fabric-ca-server-config/ca."+domainName+"-cert.pem")
	peerEnvironment = append(peerEnvironment, "FABRIC_CA_SERVER_CA_KEYFILE=/etc/hyperledger/fabric-ca-server-config/"+strings.ToUpper(orgname)+"_PRIVATE_KEY")
	peerEnvironment = append(peerEnvironment, "FABRIC_CA_SERVER_TLS_CERTFILE=/etc/hyperledger/fabric-ca-server-config/ca."+domainName+"-cert.pem")
	peerEnvironment = append(peerEnvironment, "FABRIC_CA_SERVER_TLS_KEYFILE=/etc/hyperledger/fabric-ca-server-config/"+strings.ToUpper(orgname)+"_PRIVATE_KEY")
	vols := make([]string, 0)
	vols = append(vols, cryptoBasePath+"/crypto-config/peerOrganizations/"+domainName+"/ca/"+":/etc/hyperledger/fabric-ca-server-config")

	var networks = make([]string, 0)
	networks = append(networks, "fabricnetwork")

	var container Container
	container.ContainerName = peerFQDN
	container.Environment = peerEnvironment
	container.Volumns = vols
	container.Networks = networks
	container.Ports = ports
	container.Extends = extnds
	return container
}
func BuildCouchDB(couchID string, ports []string) Container {
	var couchContainer Container
	couchContainer.ContainerName = couchID
	extnds := make(map[string]string)
	extnds["file"] = "base.yaml"
	extnds["service"] = "couchdb"
	couchContainer.Extends = extnds
	var networks = make([]string, 0)
	networks = append(networks, "fabricnetwork")

	couchContainer.Networks = networks

	couchContainer.Ports = ports
	return couchContainer
}
func BuildBaseImage(addCA bool) ServiceConfig {
	var peerbase Container
	peerEnvironment := make([]string, 0)
	peerEnvironment = append(peerEnvironment, "CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock")
	peerEnvironment = append(peerEnvironment, "CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=bc_fabricnetwork")
	peerEnvironment = append(peerEnvironment, "CORE_LOGGING_LEVEL=DEBUG")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_TLS_ENABLED=true")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_ENDORSER_ENABLED=true")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_GOSSIP_USELEADERELECTION=true")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_GOSSIP_ORGLEADER=false")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_PROFILE_ENABLED=true")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key")
	peerEnvironment = append(peerEnvironment, "CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt")

	peerbase.Image = "hyperledger/fabric-peer:${IMAGE_TAG}"
	peerbase.Environment = peerEnvironment
	peerbase.WorkingDir = "/opt/gopath/src/github.com/hyperledger/fabric/peer"
	peerbase.Command = "peer node start"
	config := make(map[string]interface{})
	config["peer"] = peerbase

	var ordererBase Container
	ordererEnvironment := make([]string, 0)
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_LOGLEVEL=debug")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_LISTENADDRESS=0.0.0.0")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_GENESISMETHOD=file")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_GENESISFILE=/var/hyperledger/orderer/genesis.block")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_LOCALMSPID=OrdererMSP")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_TLS_ENABLED=true")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_GENERAL_TLS_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_KAFKA_RETRY_SHORTINTERVAL=1s")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_KAFKA_RETRY_SHORTTOTAL=30s")
	ordererEnvironment = append(ordererEnvironment, "ORDERER_KAFKA_VERBOSE=true")

	ordererBase.Image = "hyperledger/fabric-orderer:${IMAGE_TAG}"
	ordererBase.Environment = ordererEnvironment
	ordererBase.WorkingDir = "/opt/gopath/src/github.com/hyperledger/fabric"
	ordererBase.Command = "orderer"
	config["orderer"] = ordererBase

	var couchDB Container
	couchDB.Image = "hyperledger/fabric-couchdb:${IMAGE_TAG}"
	config["couchdb"] = couchDB

	if addCA == true {
		var ca Container
		ca.Image = "hyperledger/fabric-ca:${IMAGE_TAG}"
		caEnvironment := make([]string, 0)
		caEnvironment = append(caEnvironment, "FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server")
		caEnvironment = append(caEnvironment, "FABRIC_CA_SERVER_TLS_ENABLED=true")
		ca.Environment = caEnvironment
		ca.Command = "sh -c 'fabric-ca-server start -b admin:adminpw -d'"
		config["ca"] = ca
	}
	var serviceConfig ServiceConfig
	serviceConfig.Version = "2"
	serviceConfig.Services = config

	return serviceConfig
}
func generatePorts(basePorts []int) map[int][]string {
	//Assuming we have 4 digit port
	portMap := make(map[int][]string)
	allGenerated := true
	offset := 0
	peerCount := 0
	for allGenerated == true {
		portMap[peerCount] = make([]string, 0)
		for _, port := range basePorts {
			allGenerated = allGenerated && (port+offset < 65000)
			if allGenerated {
				mapDef := fmt.Sprintf("%d:%d", port+offset, port)
				portMap[peerCount] = append(portMap[peerCount], mapDef)

			} else {
				break
			}
		}
		peerCount++
		offset += 1000
	}
	//fmt.Printf("%v\n", portMap)
	return portMap
}
