# Fractal
[![Build Status](https://travis-ci.org/fractalplatform/fractal.svg?branch=master)](https://travis-ci.org/fractalplatform/fractal)
[![GoDoc](https://godoc.org/github.com/fractalplatform/fractal?status.svg)](https://godoc.org/github.com/fractalplatform/fractal)
[![Coverage Status](https://coveralls.io/repos/github/fractalplatform/fractal/badge.svg?branch=master)](https://coveralls.io/github/fractalplatform/fractal?branch=master)
[![GitHub](https://img.shields.io/github/license/fractalplatform/fractal.svg)](LICENSE)

Welcome to the Fractal source code repository!

## What is Fractal?
Fractal is a high-level blockchain framework that can implement the issuance, circulation, and dividends of tokens efficiently and reliably. Fractal can also steadily implement various community governance functions with voting as the core and foundation. These functions are the foundation for building the token economy of future.

home page:  https://fractalproject.com/


## Executables

The fractal project comes with several wrappers/executables found in the `cmd` directory.

| Command    | Description |
|:----------:|-------------|
| **`ft`** | Our main fractal CLI client. It is the entry point into the fractal network (main-, test- or private net),  It can be used by other processes as a gateway into the fractal network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `ft -h` and the [Command Line Options](https://github.com/fractalplatform/fractal/wiki/Command-Line-Options) for command line options. |
| **`ftfinder`** | ftfinder is a fractal node discoverer.`ftfinder -h ` and the [Command Line Options](https://github.com/fractalplatform/fractal/wiki/Command-Line-Options) for command line options. |


## Getting Started
The following instructions overview the process of getting the code, building it, and start node.

### Getting the code
To download all of the code:

`git clone https://github.com/fractalplatform/fractal`

### Setting up a build/development environment

Install latest distribution of [Go](https://golang.org/) if you don't have it already. (go version >= go1.10  )

Currently supports the following operating systems: 

* Ubuntu 16.04
* Ubuntu 18.04
* MacOS Darwin 10.12 and higher

### Build Fractal

`make all`

more information see: [Installing Fractal](https://github.com/fractalplatform/fractal/wiki/Build-Fractal)


### Running a node

To run  `./ft ` , you can run your own FT instance.

`$ ft `

Join the fractal main network see: [Main Network](https://github.com/fractalplatform/fractal/wiki/Main-Network)

Join the fractal test network see: [Test Network](https://github.com/fractalplatform/fractal/wiki/Test-Network)

Operating a private network see:[Private Network](https://github.com/fractalplatform/fractal/wiki/Private-Network)

## Resources

[Fractal Official Website](https://fractalproject.com/)

[Fractal Blog](https://fractalproject.com/blog.html)


More Documentation see [the Fractal wiki](https://github.com/fractalplatform/fractal/wiki)

## License
Fractal is distributed under the terms of the [GPLv3 License](./License).
