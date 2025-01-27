### Blockchain Application

Welcome to the Blockchain Application project! This repository contains the implementation of a blockchain application using the Libp2p library.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [License](#license)

## Introduction

This project demonstrates a simple blockchain application built with the Libp2p networking library. It showcases how to create a peer-to-peer network, manage blockchain data, and ensure secure transactions.

## Features

- Peer-to-peer networking with Libp2p
- Blockchain data structure
- Secure transactions
- Consensus algorithm implementation
- Basic command-line interface
- Messaging capability

## Installation

To get started with the project, follow these steps:

1. Clone the repository:
    ```sh
    git clone https://github.com/reddy-j-harshith/Blockchain
    ```
2. Navigate to the project directory:
    ```sh
    cd blockchain
    ```
3. Install the dependencies:
    ```sh
    go mod init Messenger
    go get github.com/libp2p/go-libp2p
    go get github.com/libp2p/go-libp2p-kad-dht
    ```

## Usage

To run the blockchain application, use the following command:
```sh
go run bootstrap.go
```

To run the blockchain application, use the following command:
```sh
go run . --peer {bootstrap address}
```

You can interact with the application through the command-line interface to create transactions, mine blocks, and view the blockchain.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
