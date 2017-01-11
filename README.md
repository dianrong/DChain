

## DChain

DChain is an Ethereum-based blockchain platform with permission control and configurable consensus mechanism.

## Features

1. Enrollment mechanism and network control
	> CA Services is build to control the participant registration and access. After user obtain the certificate, one can start geth on the local and send the connect request to other peers.
 
2. Permission control
	> Different nodes assign a different account, the account has designated authority level. Each node can finely control what things can do by the permission control.

3. Configurable consensus mechanism
	> Use the sophisticated POS, PBFT and other consensus algorithm to replace the POW algorithms in order to achieve effective and efficient, to reduce the consumption of computing resources, increase the transaction throughput. The mechanism can be specified by the “common/properties.yaml”.


## Features in development

1. Confidential transaction
	> For consortium and private blockchain, confidentiality is concerned for counter parties to prevent other participants from knowing  details about the transaction. Confidentiality is a critical feature for industry.

2. Blockchain middleware
	> Blockchain middleware is designed to shield the technical details of the blockchain, providing reliable and easy to use interface for system development to the customer. Developer can develop their own applications rapidly by calling the interface provided by the application layer, including configure, monitoring, data analysis, blockchain browsing and other functions.


## Building the source

You can install them using your favourite package manager.
Once the dependencies are installed, run

'' make geth
'' make caserver
or, to build the full suite of utilities:

'' make all


## Quickstart

There must be only one caserver in your blockchain network. After the caserver is startup, update the IP of 
'' caserver : address : "10.9.22.187" 
in “common/properties.yaml” in order to join the corresponding network.

Choose the proper consensus mechanism 
'' consensus : algorithm : "POW"
in “common/properties.yaml”. Then use the command geth as Ethereum.

All the peers under the same network must use the same properties.yaml.


## Contribution

Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet.


## License

The DChain project is licensed under the
[GNU General Public License v3.0], also included
in our repository in the `COPYING` file.

