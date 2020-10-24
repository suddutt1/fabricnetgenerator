# Hyperledger Fabric Network Generator
### This tool generates hyperledger fabric v1.x network related files to spwan a network quickly

Hyperledger Fabric Network Generator
--------------------------------------------
This is a simple tool to generate a set of scripts to spawn a docker-compose based hyperledger fabric network in a linux machine. Tool takes a simple JSON as input specifying any fabric network and generates a set of shell scripts, docker-compose.yaml file to start a network, create and join channels, install and update chain codes. Tool also generates a README file to assit new users with squence of scripts to runm in order to bring a hyperledger fabric network up and running.
The saves time for a developer to build a network without making any mistake in creating configtx.yaml, crypto-config.yaml. docker-compose files etc, those are required by a hyperledger fabric network to start.
Moreover this tools helps developers to concentrate more on the smart contract writing and developing other integration parts rather than concentrating on the infrastructure part. 
The future versions of this tool is aimed to support multi-vm , K8S complient and docker swarm compliant network configuration files and scripts generation. 



## Updates 
#### Oct 25,2020: Support for 2.2.0 is completed. Testing pending.
#### Oct 10,2020: Started working to support  2.2.0 with Raft
#### July 30, 2019: Updated support for 1.4.2 with Raft
#### July 20, 2019: Stopped support for verions below Fabric version 1.4.
#### March 14, 2019: Added extra_hosts attribute same entries 
#### Dec 28,2018: Fixed issues with Fabric version 1.3 . Updated code for some new features like pre-generated user lookup and ca-affiliate-add shell scripts 
#### Dec 25,2018: Version 1.3 issues are resolved. Now you may generate and operate a hyperledger version 1.3 network
#### Dec 24,2018: Custom affiliation scripts added.Fixed other CA related issues  
#### Dec 03,2018: Anchor peer update included.  
#### Nov 28,2018: Updated to support Fabric 1.3 based network with solo anf kafka orderer. Support for Fabric 1.2 is skipped for now.
#### Nov 25,2018: Fixed issue on kafka configuration . Added a feature to generate a base chaincode
#### Nov 19,2018 : Fix issue for networks generated in ECS , Alibaba cloud environment
#### June 30, 2018: Updated a version comapatibility map system so that it can support fabric version 1.0.0, 1.1.0, 1.0.4.  
#### June 12, 2018: Added the option to generate ports starting from an input numnber
1. Refer to the startPort entry in the network-config.json
2. Tested for solo. Need to be tested for kafka based orderer options. 
3. The port numbers generated are not continous 
#### April 8, 2018: Added documentation for running the chain code after installation 
#### March 9, 2018 : Moved to HLF Version 1.1.0-rc1
#### December 25, 2017 : Added kafka option for HA orderers


## Installation  ( From Source )
1. Clone this source code
2. Build using 
    ```sh
    cd <path to source code directory>
    go get gopkg.in/yaml.v2
    go build
    ```
3. Install using  the following commands ( Make sure that GOBIN environment variable is set and your PATH contains GOBIN in it)
    ```sh
    cd <path to source code directory>
    go install
    ```
4. Create a network-config.json ( Refer to the example given in the respository).
5. Generate the scripts and other configs
    ```sh
    fabricnetgen <path to the network-config json file name>
    
 
     ```
6. Follow the instructions generated in README.txt file.

## Installation  ( Binary )
1. Download latest fabricnetgen from the releases tab
2. Change the permission to make it an executable 
 ```sh
    chmod a+x fabricnetgen
 ```  
3. Put the fabricnetgen some where so that it is in PATH  
4. Create a network-config.json ( Refer to the example given in the respository).
5. Generate the scripts and other configs
    ```sh
    fabricnetgen <path to the network-config json file name>
    
 
     ```
6. Follow the instructions generated in README.txt file.

