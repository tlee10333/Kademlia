# Kademlia
DHT Algorithm PoC



Node.go

Data Structures

Node consistents of:

1.  the Server Node ID (int)

2. KV pair where:
    key = hashed int of the string key
    value= string
3. RoutingTable: A Routing Table


Routing_table.go


NodeInfo Cosistents of:
0. ID node ID int
1. Addr = net.UDPAddr type (IP, Port, Zone)
2. Timestamp = int64 (automatically set whenever we add or interact with the node)

Different Types of Commands

4 Protocols

1. Ping
2. Store
3. Find_value
4. Find_node



Good demo/test for find_value

Server 1: 8001

distance between 1 and 4: 5 (0101)
distance between 1 and 12: 13 (1101)
distance between 1 and 13: 12 (1100)
distance between 1 and 15: 14 (1110) //Gonna be ignored


Server 4: 8004
distance between 4 and 15: 11 (1011)

Server 12 8012
distance between 12 and 13: 1 (0001)

Server 13 8013
Server 15 8015

