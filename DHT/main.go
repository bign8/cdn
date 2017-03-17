package main

import "fmt"

// DHT Distributed Hash Table
type DHT struct {
	// Id -> Node
	nodeMap map[int]Node
}

// NODE A single node (server) on the DHT
type Node struct {
	id      int
	start   int
	stop    int
	dataMap map[int]string
}

// toString method
func (n *Node) String() string {
	return fmt.Sprintf("ID: %d, Start: %d, Stop: %d\n", n.id, n.start, n.stop)
}

// Add piece of data to nodes map
func (n *Node) insertNewElement(value string, valueHash int) {
	n.dataMap[valueHash] = value
}

// Join a node to a DHT
func (dht *DHT) join(id int, start int, stop int) {
	n := &Node{
		id:      id,
		start:   start,
		stop:    stop,
		dataMap: make(map[int]string),
	}
	dht.nodeMap[id] = *n
}

// Insert new piece of data into DHT
func (dht *DHT) insertNewElement(value string, valueHash int) {
	for _, n := range dht.nodeMap {
		if n.start < valueHash && n.stop > valueHash {
			n.insertNewElement(value, valueHash)
		}
	}
	// TODO what happens if we walk all the away around and don't find hashValue?
}

func (dht *DHT) buildTable(data []string) {
	// dataSize := len(data)
	// hash := adler32.New()
	dataSize := 100
	numNodes := 3 //TODO: decide how to create initial number of nodes
	start := 0

	//add nodes to table
	// TODO: Better assignment of spaces to cover for nodes
	for i := 0; i < numNodes; i++ {
		stop := start + dataSize/numNodes
		dht.join(i, start, stop)
		start = stop
	}

	// Create hashes for each data element
	for _, v := range data {
		// hv, _ := hash.Write([]byte(v))
		hv := simpleASCIIHash(v, dataSize)
		dht.insertNewElement(v, hv)
	}

}

// Sum Ascii values in given string
func sumChars(input string) int {
	var sum = 0
	for _, elem := range input {
		sum += int(elem)
	}
	return sum
}

// Create simple hash of string by summing Ascii values then mod
// by capacity
func simpleASCIIHash(input string, capacity int) int {
	hash := sumChars(input)
	return hash % capacity
}

func main() {

	l := []string{"Lisa Peters", "Nate Woods", "Jani Rounds", "Noah Visscher", "Moby Thomas", "Leila Dog", "Shoka Cat Devil"}

	// Create hashtable
	dht := &DHT{
		nodeMap: make(map[int]Node),
	}
	dht.buildTable(l)

	for _, n := range dht.nodeMap {
		fmt.Println(n)
	}

}
