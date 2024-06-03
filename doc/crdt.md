# Mutual agreement in messages order and nodes management in a multi-user chat.

## Operation based CRDTs
Propagation of commutative operations rather than whole node data :

- add or remove a room.
- add, update or remove a message from a room (messages order is chosen based on sending date).
- add, remove or remove a node from a given room.

## Synchronisation strategy
The nodes are connected in a fully meshed network (each node from a room has an open TCP connection to each node of the room)
For now, each node propagates its operations to all other nodes : 

- a new node comes in the room (the entry point node first forwards the add node operation to the other nodes).
- a message is added, updated or removed in the room.
- a node leaves the room.