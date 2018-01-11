package main

import "fmt"

func GenerateNetworkItems(configBytes []byte, baseOutputPath string) {
	if !GenerateConfigTxGen(configBytes, baseOutputPath+"/configtx.yaml") {
		fmt.Errorf("Error in generation of configtx.yaml")
	}
	fmt.Println("configtx.yaml generated ...")
	if !GenerateCrytoConfig(configBytes, baseOutputPath+"/crypto-config.yaml") {
		fmt.Errorf("Error in generation of crypto-config.yaml")
		return
	}

	fmt.Println("crypto-config.yaml generated....")
	if !GenerateDockerFiles(configBytes, baseOutputPath) {
		fmt.Errorf("Error in generating the docker files")
	}
	fmt.Println("Generated docker-compose.yaml ..")
	fmt.Println("setpeers.sh generation in progress ....")
	if !GenerateSetPeer(configBytes, baseOutputPath+"/setpeer.sh") {
		fmt.Errorf("Error in generating the setpeer.sh")
	}
	fmt.Println("generateartifacts.sh generation in progress ....")
	if !GenerateGenerateArtifactsScript(configBytes, baseOutputPath+"/generateartifacts.sh") {
		fmt.Errorf("Error in generating the generateartifacts.sh")
	}

	fmt.Println("buildandjoinchannel.sh generation in progress ....")
	if !GenerateBuildAndJoinChannelScript(configBytes, baseOutputPath+"/buildandjoinchannel.sh") {
		fmt.Println("Error in generating the buildandjoinchannel.sh")
	}
	fmt.Println("Generating misc scripts ....")
	if !GenerateOtherScripts(baseOutputPath + "/") {
		fmt.Println("Error in generating misc scripts")
	}
	fmt.Println("Generating chaincode related scripts ....")
	if !GenerateChainCodeScripts(configBytes, baseOutputPath+"/") {
		fmt.Println("Error in generating chain code related scripts")
	}
	fmt.Println("Generating README.txt")
	GenerateReadme(configBytes, baseOutputPath+"/")
}
