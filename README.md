# OEXChain

[![Build Status](https://travis-ci.org/oexplatform/fractal.svg?branch=master)](https://travis-ci.org/oexplatform/fractal)
[![GoDoc](https://godoc.org/github.com/oexplatform/fractal?status.svg)](https://godoc.org/github.com/oexplatform/fractal)
[![Coverage Status](https://coveralls.io/repos/github/oexplatform/fractal/badge.svg?branch=master)](https://coveralls.io/github/oexplatform/fractal?branch=master)
[![GitHub](https://img.shields.io/github/license/oexplatform/fractal.svg)](LICENSE)

Welcome to the OEXChain source code repository!

## What is OEX chain?

OEX chain is a high-level blockchain framework that can implement the issuance, circulation, and dividends of tokens efficiently and reliably. OEX chain can also steadily implement various community governance functions with voting as the core and foundation. These functions are the foundation for building the token economy of future.

home page: http://oex.com/

## Executables

The OEX chain project comes with several wrappers/executables found in the `cmd` directory.

|    Command     | Description                                                                                                                                                                                                                                                                                                                                                                                               |
| :------------: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
|    **`oex`**    | Our main OEX CLI client. It is the entry point into the OEX network (main-, test- or private net), It can be used by other processes as a gateway into the oex network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `oex -h` and the [Command Line Options](https://github.com/oexplatform/oexchain/wiki/Command-Line-Options) for command line options. |
| **`oexfinder`** | oexfinder is a OEXChain node discoverer.`oexfinder -h` and the [Command Line Options](https://github.com/oexplatform/oexchain/wiki/Command-Line-Options) for command line options.                                                                                                                                                                                                                        |

## Getting Started

The following instructions overview the process of getting the code, building it, and start node.

### How to create first account

You can use [OEX wallet](https://m.oex.com/download) to create first account.

### Getting the code

To download all of the code:

`git clone https://github.com/oexplatform/oexchain`

### Setting up a build/development environment

Install latest distribution of [Go](https://golang.org/) if you don't have it already. (go version >= go1.10 )

Currently supports the following operating systems:

- Ubuntu 16.04
- Ubuntu 18.04
- MacOS Darwin 10.12 and higher

### Build OEXChain

`make all`

more information see: [Installing OEXChain](https://github.com/oexplatform/oexchain/wiki/Build-OEX)

### Running a node

To run `./oex` , you can run your own OEX instance.

`$ oex`

Join the OEXChain main network see: [Main Network](https://github.com/oexplatform/oexchain/wiki/Main-Network)

Join the OEXChain test network see: [Test Network](https://github.com/oexplatform/oexchain/wiki/Test-Network)

Operating a private network see:[Private Network](https://github.com/oexplatform/oexchain/wiki/Private-Network)

## Resources

[OEX Official Website](http://oex.com/)

[OEX Blog](http://oex.com/blog.html)

More Documentation see [the OEXChain wiki](https://github.com/oexplatform/oexchain/wiki)

## License

OEXChain is distributed under the terms of the [GPLv3 License](./License).
