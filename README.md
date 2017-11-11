# Hyperledger Fabric Network Generator
### This tool generates hyperledger fabric v1.x network related files to spwan a network quickly
Hyperledger Fabric Network Generator

## TODO
1. Need to document the steps 
2. Need to add some command line option to generate input network json file
3. Add instructions at the end of generation 

## Tool download

```sh
export VERSION=1.0.4
export ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')
#Set MARCH variable i.e ppc64le,s390x,x86_64,i386
MARCH=`uname -m`
echo "===> Downloading platform binaries"
curl https://nexus.hyperledger.org/content/repositories/releases/org/hyperledger/fabric/hyperledger-fabric/${ARCH}-${VERSION}/hyperledger-fabric-${ARCH}-${VERSION}.tar.gz | tar xz



```
