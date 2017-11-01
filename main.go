package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	args := os.Args[1:]
	fmt.Printf("Starting the application.... \n")
	fmt.Printf("Reading the input .... %v\n", args[0])
	configBytes, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Errorf("Error in reading input json")
		return
	}
	if !GenerateConfigTxGen(configBytes, "./configtx.yaml") {
		fmt.Errorf("Error in generation of configtx.yaml")
	}
	fmt.Println("configtx.yaml generated ...")
	if !GenerateCrytoConfig(configBytes, "./crypto-config.yaml") {
		fmt.Errorf("Error in generation of crypto-config.yaml")
		return
	}

	fmt.Println("crypto-config.yaml generated....")
	if !GenerateDockerFiles(configBytes, ".") {
		fmt.Errorf("Error in generating the docker files")
	}
	fmt.Println("Generated docker-compose.yaml ..")
	fmt.Println("setpeers.sh generation in progress ....")
	if !GenerateSetPeer(configBytes, "./setpeer.sh") {
		fmt.Errorf("Error in generating the setpeer.sh")
	}
	fmt.Println("generateartifacts.sh generation in progress ....")
	if !GenerateGenerateArtifactsScript(configBytes, "./generateartifacts.sh") {
		fmt.Errorf("Error in generating the generateartifacts.sh")
	}

	fmt.Println("buildandjoinchannel.sh generation in progress ....")
	if !GenerateBuildAndJoinChannelScript(configBytes, "./buildandjoinchannel.sh") {
		fmt.Errorf("Error in generating the buildandjoinchannel.sh")
	}
	fmt.Println("Generating misc scripts ....")
	if !GenerateOtherScripts("./") {
		fmt.Errorf("Error in generating misc scripts")
	}
}
