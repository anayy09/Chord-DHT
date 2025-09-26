package chord

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

// Constants
const (
	mBits = 160 // Real Chord uses 160
)

func Hash(s string) *big.Int {
	h := sha1.Sum([]byte(s))
	return new(big.Int).SetBytes(h[:])
}

// NodeInfo holds PID and ID
type NodeInfo struct {
	PID *actor.PID
	ID  *big.Int
}

// Messages
type Join struct{ ExistingNode *actor.PID } // New node sends this to itself with existing PID
type FindSuccessor struct {
	Key       *big.Int
	Requester *actor.PID
}
type FoundSuccessor struct{ Successor *NodeInfo }
type GetPredecessor struct{}
type GetPredecessorResponse struct{ Predecessor *NodeInfo }
type Notify struct{ Node *NodeInfo }
type Stabilize struct{}
type Lookup struct{ Key *big.Int }
type Store struct {
	Key   *big.Int
	Value string
}
type GetValue struct {
	Key       *big.Int
	Requester *actor.PID
}
type GetValueResponse struct {
	Key   *big.Int
	Value string
	Found bool
}
type GetSuccessor struct{}
type GetSuccessorResponse struct{ Successor *NodeInfo }

// ChordNode actor
type ChordNode struct {
	self          *NodeInfo
	successor     *NodeInfo
	predecessor   *NodeInfo
	fingers       []*NodeInfo
	nextFinger    int
	context       *actor.RootContext
	storage       map[string]string
	successorList []*NodeInfo
}

func NewChordNode(id *big.Int, context *actor.RootContext) actor.Producer {
	return func() actor.Actor {
		self := &NodeInfo{ID: id}
		node := &ChordNode{
			self:          self,
			fingers:       make([]*NodeInfo, mBits),
			context:       context,
			nextFinger:    0,
			storage:       make(map[string]string),
			successorList: make([]*NodeInfo, 0, 5),
		}
		return node
	}
}

func (n *ChordNode) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		n.self.PID = ctx.Self()
		if n.successor == nil {
			n.successor = n.self // Bootstrap sets successor to self
		}
		// Start periodic stabilization
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					n.context.Send(n.self.PID, &Stabilize{})
				}
			}
		}()
	case *Join:
		// Initiate join: Find my successor via existing node
		ctx.Request(msg.ExistingNode, &FindSuccessor{Key: n.self.ID, Requester: ctx.Self()})
	case *FindSuccessor:
		if n.successor == nil {
			ctx.Send(msg.Requester, &FoundSuccessor{Successor: n.self})
			return
		}
		if between(n.self.ID, msg.Key, n.successor.ID) {
			ctx.Send(msg.Requester, &FoundSuccessor{Successor: n.successor})
		} else {
			next := n.closestPrecedingNode(msg.Key)
			if next != nil {
				ctx.Request(next.PID, msg)
			} else {
				ctx.Send(msg.Requester, &FoundSuccessor{Successor: n.successor})
			}
		}
	case *FoundSuccessor:
		// Set my successor after join
		n.successor = msg.Successor
		fmt.Printf("Node %s set successor to %s\n", n.self.ID, msg.Successor.ID)
	case *GetPredecessor:
		ctx.Respond(&GetPredecessorResponse{Predecessor: n.predecessor})
	case *GetPredecessorResponse:
		if msg.Predecessor != nil && between(n.self.ID, msg.Predecessor.ID, n.successor.ID) {
			n.successor = msg.Predecessor
		}
		n.context.Send(n.successor.PID, &Notify{Node: n.self})
		// Update successor list starting with current successor
		n.successorList = []*NodeInfo{n.successor}
	case *Notify:
		if n.predecessor == nil || between(n.predecessor.ID, msg.Node.ID, n.self.ID) {
			n.predecessor = msg.Node
		}
	case *Stabilize:
		if n.successor != nil {
			n.context.RequestWithCustomSender(n.successor.PID, &GetPredecessor{}, ctx.Self())
			n.context.RequestWithCustomSender(n.successor.PID, &GetSuccessor{}, ctx.Self())
			n.fixFingers()
		}
	case *Lookup:
		succ := n.findSuccessor(msg.Key)
		if succ != nil {
			n.context.Send(succ.PID, &GetValue{Key: msg.Key, Requester: ctx.Self()})
		} else {
			fmt.Println("Lookup failed: no successor")
		}
	case *Store:
		succ := n.findSuccessor(msg.Key)
		if succ.PID == n.self.PID {
			n.storage[msg.Key.String()] = msg.Value
			fmt.Printf("Stored key %s with value %s at node %s\n", msg.Key, msg.Value, n.self.ID)
		} else {
			n.context.Send(succ.PID, msg)
		}
	case *GetValue:
		if value, found := n.storage[msg.Key.String()]; found {
			n.context.Send(msg.Requester, &GetValueResponse{Key: msg.Key, Value: value, Found: true})
		} else {
			n.context.Send(msg.Requester, &GetValueResponse{Key: msg.Key, Value: "", Found: false})
		}
	case *GetValueResponse:
		if msg.Found {
			fmt.Printf("Lookup result: Key %s has value %s\n", msg.Key, msg.Value)
		} else {
			fmt.Printf("Lookup result: Key %s not found\n", msg.Key)
		}
	case *GetSuccessor:
		ctx.Respond(&GetSuccessorResponse{Successor: n.successor})
	case *GetSuccessorResponse:
		// Update successor list
		n.successorList = append(n.successorList, msg.Successor)
		if len(n.successorList) > 5 {
			n.successorList = n.successorList[:5]
		}
	}
}

func (n *ChordNode) closestPrecedingNode(key *big.Int) *NodeInfo {
	for i := mBits - 1; i >= 0; i-- {
		f := n.fingers[i]
		if f != nil && between(n.self.ID, f.ID, key) {
			return f
		}
	}
	return n.successor
}

func (n *ChordNode) findSuccessor(key *big.Int) *NodeInfo {
	if n.successor == nil {
		return nil
	}
	if between(n.self.ID, key, n.successor.ID) {
		return n.successor
	}
	next := n.closestPrecedingNode(key)
	if next == nil || next.PID == n.self.PID {
		return n.successor
	}
	// In full impl, request remotely; here return next for simulation
	return next
}

func (n *ChordNode) fixFingers() {
	n.nextFinger = (n.nextFinger + 1) % mBits
	fingerKey := new(big.Int).Add(n.self.ID, new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(n.nextFinger)), nil))
	mod := new(big.Int).Exp(big.NewInt(2), big.NewInt(mBits), nil)
	fingerKey.Mod(fingerKey, mod)
	n.fingers[n.nextFinger] = n.findSuccessor(fingerKey)
}

func between(left, elem, right *big.Int) bool {
	if left == nil || elem == nil || right == nil {
		return false
	}
	if right.Cmp(left) > 0 {
		return elem.Cmp(left) > 0 && elem.Cmp(right) <= 0
	}
	return elem.Cmp(left) > 0 || elem.Cmp(right) <= 0
}
