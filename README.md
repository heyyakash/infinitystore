# InfinityStore Documentation 📚
InfinityStore is a distributed key-value store built using the [Raft](https://developer.hashicorp.com/consul/docs/architecture/consensus) Consensus algorithm for high availability and consistency. This documentation provides an overview of how to run the application and interact with its API.
An easier explaination of raft can be found [here](https://yusufs.medium.com/creating-distributed-kv-database-by-implementing-raft-consensus-using-golang-d0884eef2e28)

# Table of Contents
1. [Running the Application 🚀](#running)
1. [HTTP API 📡](#http)
1.1. Set
1.2. Get
1.3. Join
2. [Dependencies 📦](#dependencies)
3. [Configuration ⚙️](#config)
4. [Example](#example)

## Running the Application 🚀 <a name="running"><a/>
To run InfinityStore, follow these steps:

### Clone the repository:

```sh
git clone https://github.com/heyyakash/infinitystore.git
cd infinitystore
```

### Build the application
```sh
go build -o store
```

### Run the store with the necessary flags:

```sh
./infinitystore --node-id=node1 --raft-addr=localhost:8000 --http-addr=:8001
```

## HTTP API 📡 <a name="http"><a/>
InfinityStore exposes several HTTP endpoints to interact with the distributed key-value store.

### Set 📥
Endpoint: /set
Method: POST
Description: Sets a key-value pair in the store.

Request Body:

```json
{
    "action" :"set"
    "key": "your_key",
    "value": "your_value"
}
```
Example:

```sh
curl -X POST http://localhost:8001/set -d '{"key":"foo", "value":"bar","action":"set"}'
curl -X POST http://localhost:8001/set -d '{"key":"foo", "value":"bar","action":"delete"}'
```

Response:
```sh
200 OK: Set successful
```
```sh
400 Bad Request: Invalid body
```
```ssh
500 Internal Server Error: Couldn't add to store
```

### Get 📤
Endpoint: /get
Method: GET
Description: Retrieves the value for a given key.

Query Parameters:
key: The key to retrieve the value for.
Example:

```sh
curl -X GET "http://localhost:8001/get?key=foo"
```
Response:
```sh
200 OK: Key and value
```
```sh
400 Bad Request: Missing key
```
```sh
404 Not Found: Key does not exist
```

### Join 🤝
Endpoint: /join
Method: POST
Description: Joins a new node to the cluster.

Query Parameters:

nodeid: The ID of the node to join.
nodeaddr: The address of the node to join.
Example:

```sh
curl -X POST "http://localhost:8001/join?nodeid=node2&nodeaddr=localhost:8002"
```
Response:
```sh
200 OK: Voter added
```
```ssh
400 Bad Request: Not a leader
```
```ssh
500 Internal Server Error: Failed to add voter
```

## Dependencies 📦 <a name="dependencies"><a/>
InfinityStore relies on several external packages:

1. github.com/gorilla/mux: Router for handling HTTP requests.
2. github.com/hashicorp/raft: Implementation of the Raft consensus protocol.
3. github.com/heyyakash/infinitystore/datastore: Custom data store implementation.
4. github.com/heyyakash/infinitystore/fsmachine: Finite state machine for Raft.
5. github.com/heyyakash/infinitystore/helper: Helper functions for response generation.
6. github.com/heyyakash/infinitystore/models: Data models used in the application.
7. github.com/heyyakash/infinitystore/raft: Raft consensus setup and management.

## Configuration ⚙️ <a name="config"><a/>
The application configuration is managed via command-line flags:

```sh
node-id: Unique identifier for the node (default: "node1").
raft-addr: Address for the Raft server (default: "localhost:8000").
http-addr: Address for the HTTP server (default: ":8001").
These flags are parsed during the initialization phase:
```

```go
func Init() {
    nodeID = flag.String("node-id", "node1", "Node ID")
    raftAddr = flag.String("raft-addr", "localhost:8000", "Raft Server Address")
    httpAddr = flag.String("http-addr", ":8001", "Http server Address")
    flag.Parse()
}
```

### Example Usage <a name="example"><a/>
For simplicity we'll create 3 instances to imitate 3 nodes on same node.
1. Create a Leader Node
   ```sh
   ./store --http-addr=:8001 --raft-addr=localhost:8000 --node-id=node-1
   ```
2. Create a follower node
   ```sh
   ./store --http-addr=:8011 --raft-addr=localhost:8001 --node-id=node-2
   ```
3. Create another follower node
   ```sh
   ./store --http-addr=:8021 --raft-addr=localhost:8020 --node-id=node-3
   ```
```Note``` :To gain majority minimum no. of available nodes should be (N+1)/2. So for a basic cluster we are going to have 3 nodes.

4. Join the Follower nodes to the Leader node.
    ```sh
    curl -X POST "http://localhost:8001/join?nodeid=node-2&nodeaddr=localhost:8010"
    curl -X POST "http://localhost:8001/join?nodeid=node-3&nodeaddr=localhost:8020"
    ```
5. Now send a set request to Leader Node
    ```sh
    curl -X POST http://localhost:8001/set -d '{"key":"foo", "value":"bar","action":"set"}'
    ```
6. Now send a get request to one of the follower 
    ```sh
    curl -X GET "http://localhost:8011/get?key=foo"
    ```











