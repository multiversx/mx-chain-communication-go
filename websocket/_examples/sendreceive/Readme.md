## WebSocket Client-Server Example

This example demonstrates a simple client-server interaction using WebSocket communication. 
The setup includes a client folder and a server folder, each containing a respective binary.
The server binary starts a WebSocket server and sends a message, while the client binary 
starts a WebSocket client and waits to receive a message.

## Usage

1. Open a terminal or command prompt.
2. Navigate to the client folder.
3. Built the client
``` bash
    go build && ./client
```
4. Open another terminal or command prompt.
5. Navigate to the server folder.
``` bash
    go build && ./server
```
7. Wait for the server to successfully send the message and the client to receive it.
8. Once the message has been sent and received successfully, the processes will end.
