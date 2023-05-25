# mx-chain-communication-go

This repository contains a Go modules for implementing different communication systems.

## Features

- WebSocket client-server implementation for communication
- Real-time data exchange between participants
- Scalable architecture for handling multiple client connections
- Customizable client/server configurations
- The `Messenger` implementation that handles the communication between MultiversX nodes

## Sections

### WebSocket communication

The web socket functionality allows real-time bidirectional communication between a client and a server. 
It provides a persistent connection that enables instant data transmission and updates.

#### Examples
The [examples](./websocket/examples) folder contains a demonstration of how to send and receive messages using the WebSocket host implemented in this repository. 
This example provides a basic usage scenario to help you understand and get started with the WebSocket functionality.

### P2P communication

The peer-to-peer communication is managed by the `Messenger` implementation, which handles both messages broadcasted through the entire network and messages sent from directly connected peers.

# Contributing
Contributions to the mx-chain-communication-go module are welcomed. If you find any issues or have suggestions for improvements,
please create a new issue or submit a pull request.

# License

This project is licensed under the [GNU License](https://github.com/multiversx/mx-chain-communication-go/blob/main/LICENSE).


