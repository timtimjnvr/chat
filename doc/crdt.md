# Mutual agreement in messages order between nodes

## Operation based CRDTs
propagation of commutative operations rather than whole node state (chat history) :
 
- add, update or remove a message from a discussion given the message id (uuid).
- messages order is chosen based on sending date.

## Synchronisation strategy (IN PROGRESS)
For now, each node propagates his operations to all other nodes.