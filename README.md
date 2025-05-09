# Kademlia
DHT Algorithm PoC

This is a basic implementation of a Distributed Hash Table (DHT), specifically using Kademlia's algorithm and protocol in Golang. Kademlia is a peer-to-peer (P2P) overlay network protocol used for decentralized lookup and storage, most famously used in BitTorrent and IPFS.

In this sample, K= 2 and bitLength = 4.
Features:

- XOR distance metric for node lookup
- K-bucket routing table with eviction policy
- Node discovery using recursive FIND_NODE
- Basic message types: PING, STORE, FIND_NODE, FIND_VALUE


To run this, first clone this repository through a terminal/command prompt

```
git clone https://github.com/yourusername/Kademlia.git
cd Kademlia
```
To run this project, you can just do:

```
go run main.go routing_table.go node.go <PORT #>
```
This spins up a Kademlia node listening on port <PORT #> with a basic terminal interface. Currently, the code has been set to generate a server ID based on the port number of the last digits of the port number, restricted by how many bits we want to work with (currently bitLength is set to 4 bits), so port 8010 will give us a server ID of 10, 9011 gives us server ID of 11, etc. The code also currently restricts to UDP communication between ports only on localhost.


By opening up multiple terminals and having multiple servers on different ports, you can see different communication between servers. 

Below are all possible commands:

- ping <port> — Send a ping to a node running on the specified port.
- store <port> <key> <value> — Store a key-value pair at a specific node.
- store_dht <key> <value> — Store a key-value pair in the distributed hash table (DHT).
- find_value <key> — Find the value associated with a key in the DHT.
- find_node <nodeID> — Find nodes closest to the given node ID.
- print_rtable — Print the routing table of the current server node.

## Sample Sequence
Below is a sample sequence between 2 servers, showcasing how ping, store, and find_value work between these two servers. 

1. START SERVER 1
```
go run main.go routing_table.go node.go 8001
```

2. START SERVER 2 (seperate terminal)
```
go run main.go routing_table.go node.go 8002

```

3. PING (From server 1)
```
ping 8002
```
If 8002 does not exist or does not respond fast enough, server 1 will delete it from their routing table if the server was in their routing table. In this example this doesn't apply, but you can delete server 2 after ping to see this in action. 

4. STORE (from server 1)
```
store 8001 hello world
```
This will store hello world KV pair onto server 1 (regardless of ID distance)

5. FIND_VALUE (from server 2)
```
find_value hello
```
This will allow server 2 to query server 1 to see if it has the value to hello. 

If you have 3+ servers, you can do `store_dht` which will automatically decide where to allocate the KV pair based on how close the key is to the server ID via XOR (K=2 for redundancy)



