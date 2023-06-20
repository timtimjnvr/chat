# Mutual agreement in messages order and nodes management in a multi-user chat.

## Operation based CRDTs
Propagation of commutative operations rather than whole node state (chat history) :
 
- add, update or remove a message from a discussion (messages order is chosen based on sending date).
- add, remove a node from a given discussion.

## Synchronisation strategy
The nodes are connected in a fully meshed network (each node from a discussion has an open TCP connection to each node of the discussion)
For now, each node propagates its operations to all other nodes : 

- a message is added, updated or removed in the discussion.
- a new node arrives in the chat (the entry point node first forwards the add node operation to the other nodes).
- a node leaves the chat.