package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {

	//flag.Usage = toolUsage
	action := "generateNetwork"
	genExample := flag.Bool("ccex", false, "Generate example chaincode")
	v2Example := flag.Bool("v2", true, "Generate samples in fabric 2.2.x compliant")
	huntUsers := flag.Bool("lu", false, "Lookup crypto-config directory for pregenerated users and generates config snippet")
	isJSONOutput := flag.Bool("json", false, "Generate config in json format")
	genNetconfg := flag.Bool("ncex", false, "Generate an example network-config.json")
	connProf := flag.Bool("cp", false, "Generate connection profile")
	flag.Parse()
	if *genExample {
		action = "generateExample"
	}
	if *genNetconfg {
		action = "generateConfExample"
	}
	if *huntUsers {
		action = "huntUsers"
	}
	if *connProf {
		action = "gen-connection-profile"
	}

	args := flag.Args()
	fmt.Printf("Starting the application.... \n")
	switch action {
	case "generateNetwork":
		if len(args) == 0 {
			flag.Usage()
			os.Exit(1)
		}
		fmt.Printf("Reading the input .... %v\n", args[0])
		configBytes, err := ioutil.ReadFile(args[0])
		if err != nil {
			fmt.Println("Error in reading input json")
			os.Exit(2)
		}

		GenerateNetworkItems(configBytes, ".")
	case "generateExample":
		if !*v2Example {
			GenerateExampleCC("v1", "./")
		} else {
			GenerateExampleCCV2("v2", "./")
		}
	case "generateConfExample":
		GenerateExampleConfig("v1", "./")
	case "huntUsers":
		cryptoDirectory := "."
		if len(flag.Args()) > 0 {
			cryptoDirectory = flag.Args()[0]
		}
		HuntCertificates(cryptoDirectory, *isJSONOutput)
	case "gen-connection-profile":
		if len(args) == 0 {
			flag.Usage()
			os.Exit(1)
		}
		nc := NewNetworkConfig(args[0])
		nc.GenerateConnectionProfile()
	default:
		flag.Usage()
	}

}

var toolUsage = func() {
	fmt.Printf("Usage : fabricnetgen [flags] <network json file >\n")
	fmt.Printf("Flags : -ex Generates an example chaincode in the current directory\n")

}
