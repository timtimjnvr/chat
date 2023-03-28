# Mutual agreement in messages order and nodes management in a multi-user chat.

## Operation based CRDTs
Propagation of commutative operations rather than whole node state (chat history) :
 
- add, update or remove a message from a discussion history given the message id (uuid).
- messages order is chosen based on sending date.

## Synchronisation strategy
For now, each node propagates his operations to all other nodes (except himself). 

Synchronisation happens when :
- a message is sent.
- a new node arrives in the chat.
- a node leaves the chat.