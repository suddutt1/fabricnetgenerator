# Hyperledger Fabric Network Generator
### This tool generates hyperledger fabric v1.x network related files to spwan a network quickly
Hyperledger Fabric Network Generator

## Installation 
1. Clone this source code
2. Build using 
    ```sh
    cd <path to source code directory>
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
6. After generating the network elements run the following commands to build and start the network. Remember that all the artifacts would be generated in the directory from where fabricnetgen command is invoked.
    ```sh
    . downloadbin.sh # One time command
    . generateartifacts.sh # One time to generate the crypto materials
    docker-compose up -d # To start the network
    docker exec -it cli bash -f ./buildandjoinchannel.sh # To build and join channel
    docker exec -it cli bash -f ./<chaincode id>_install.sh # To install the chain code
    
    ```

## TODO
1. Need to document the steps 
2. Need to add some command line option to generate input network json file
3. Add instructions at the end of generation 
