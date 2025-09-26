package main

import (
	"fmt"
	"os"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"

	"chord_example/chord"
)

func main() {
	system := actor.NewActorSystem()
	hostname, _ := os.Hostname()
	config := remote.Configure(hostname, 8080)
	remoter := remote.NewRemote(system, config)
	remoter.Start()

	// Create bootstrap node
	bootstrapID := chord.Hash("bootstrap")
	bootstrapProps := actor.PropsFromProducer(chord.NewChordNode(bootstrapID, system.Root))
	bootstrapPID := system.Root.Spawn(bootstrapProps)

	// Create new node and initiate join
	newNodeID := chord.Hash("node42")
	newNodeProps := actor.PropsFromProducer(chord.NewChordNode(newNodeID, system.Root))
	newNodePID := system.Root.Spawn(newNodeProps)
	system.Root.Send(newNodePID, &chord.Join{ExistingNode: bootstrapPID})

	// Create additional nodes
	node100ID := chord.Hash("node100")
	node100Props := actor.PropsFromProducer(chord.NewChordNode(node100ID, system.Root))
	node100PID := system.Root.Spawn(node100Props)
	system.Root.Send(node100PID, &chord.Join{ExistingNode: bootstrapPID})

	node200ID := chord.Hash("node200")
	node200Props := actor.PropsFromProducer(chord.NewChordNode(node200ID, system.Root))
	node200PID := system.Root.Spawn(node200Props)
	system.Root.Send(node200PID, &chord.Join{ExistingNode: bootstrapPID})

	// Simulate stabilization
	time.Sleep(time.Second)
	system.Root.Send(newNodePID, &chord.Stabilize{})
	system.Root.Send(node100PID, &chord.Stabilize{})
	system.Root.Send(node200PID, &chord.Stabilize{})

	// Store some values
	system.Root.Send(bootstrapPID, &chord.Store{Key: chord.Hash("key1"), Value: "value1"})
	system.Root.Send(newNodePID, &chord.Store{Key: chord.Hash("key2"), Value: "value2"})
	system.Root.Send(node100PID, &chord.Store{Key: chord.Hash("key3"), Value: "value3"})

	time.Sleep(time.Second)

	// Lookup multiple keys
	keys := []string{"key1", "key2", "key3", "key4"}
	for _, k := range keys {
		system.Root.Send(newNodePID, &chord.Lookup{Key: chord.Hash(k)})
	}

	// Simulate node failure
	time.Sleep(2 * time.Second)
	system.Root.Stop(node100PID)
	fmt.Println("Stopped node100")

	time.Sleep(time.Second)
	// Lookup again after failure
	for _, k := range keys {
		system.Root.Send(newNodePID, &chord.Lookup{Key: chord.Hash(k)})
	}

	time.Sleep(5 * time.Second) // Wait for operations to complete
}
