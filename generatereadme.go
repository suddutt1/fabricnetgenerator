package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func GenerateReadme(config []byte, path string) {

	var buffer bytes.Buffer
	buffer.WriteString("## Installation steps ")
	buffer.WriteString("Unzip and place the contents generated in a directory. \n")
	buffer.WriteString("The directory is referred as <network>.  \n")
	buffer.WriteString("All the commands must be executed from  <network> directory.  \n")
	buffer.WriteString("## First time setup instructions ( START) \n")
	buffer.WriteString("## First time setup. Run the following commands \n")

	buffer.WriteString(" 1. Download the binaries \n")
	buffer.WriteString("```sh \n")
	buffer.WriteString(" . ./downloadbin.sh \n")
	buffer.WriteString("``` \n")
	buffer.WriteString("\n 2. To Generate the cryto config and other configurations\n ")
	buffer.WriteString("```sh \n")
	buffer.WriteString(" . ./generateartifacts.sh \n\n")
	buffer.WriteString("``` \n")
	buffer.WriteString("\n 3. Create the chain code directiory.\n")

	configMap := make(map[string]interface{})
	//chain code
	json.Unmarshal(config, &configMap)
	ccList := configMap["chaincodes"].([]interface{})
	buffer.WriteString("```sh \n")
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})
		buffer.WriteString("\n  export NETWORKDIR=`pwd` ")
		buffer.WriteString(fmt.Sprintf("\n  mkdir -p chaincode/%s ", ccDetailsMap["src"]))
		buffer.WriteString(fmt.Sprintf("\n  cd chaincode/%s ", ccDetailsMap["src"]))
		buffer.WriteString("\n  go mod vendor")
		buffer.WriteString("\n  cd $NETWORKDIR \n\n")

	}
	buffer.WriteString("``` \n")
	buffer.WriteString("\n 4. Copy the chain code files in the respectivive directories \n")
	buffer.WriteString("\n 5. Start the netowrk  \n\n")
	buffer.WriteString("```sh \n")
	buffer.WriteString("  . setenv.sh \n")
	buffer.WriteString("  docker-compose up -d \n")
	buffer.WriteString("``` \n")
	buffer.WriteString("\n 6. Build and join channel. Make sure that network is running \n\n")
	buffer.WriteString("```sh \n")
	buffer.WriteString("   docker exec -it cli bash -e ./buildandjoinchannel.sh \n\n")
	buffer.WriteString("``` \n")
	buffer.WriteString("\n 7. Install and intantiate the chain codes \n")
	buffer.WriteString("```sh \n")
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})

		buffer.WriteString(fmt.Sprintf("  docker exec -it cli bash -e  ./%s_install.sh\n", ccDetailsMap["ccid"].(string)))
	}
	buffer.WriteString("``` \n")
	buffer.WriteString("##  First time setup instructions ( END) \n")
	buffer.WriteString("\n\n ## When chain code is modified ( START) \n")
	buffer.WriteString(" To update the chain code , first update the chain code source files in the above mentioned directory.\n" + "Then run the following commands as appropriate\n\n")
	buffer.WriteString("```sh \n")
	for _, ccDetails := range ccList {
		ccDetailsMap := ccDetails.(map[string]interface{})
		buffer.WriteString(fmt.Sprintf("  docker exec -it cli bash -e  ./%s_update.sh <version>\n", ccDetailsMap["ccid"].(string)))
	}
	buffer.WriteString("``` \n")
	buffer.WriteString("## When chain code is modified ( END) \n")
	buffer.WriteString("\n\n## To bring up an existing network ( START) \n")
	buffer.WriteString("``` \n")
	buffer.WriteString("  . setenv.sh \n")
	buffer.WriteString("  docker-compose up -d \n")
	buffer.WriteString("``` \n")
	buffer.WriteString("## To bring up an existing network ( END) \n")
	buffer.WriteString("\n\n##To destory  an existing network ( START) \n")
	buffer.WriteString("``` \n")
	buffer.WriteString("  . setenv.sh \n")
	buffer.WriteString("  docker-compose down \n")
	buffer.WriteString("``` \n")
	buffer.WriteString(" If you are stoping a network using the above commands , " +
		"\n to start the network again , you have to execute steps 2,5,6,7 of the first time setup.\n ")

	buffer.WriteString("\n\n## To destory  an existing network ( END) \n")
	ioutil.WriteFile(path+"/README.md", buffer.Bytes(), 0666)

}
