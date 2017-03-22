package DHT

import (
	"log"
	"math"
	"sort"
)

//SimplisticDHT ... Simplistic hashtable, assumes nodes (cdns) hold data in linear order, simply decrement close to query hash to find owner
type SimplisticDHT struct {
	DataMap    map[int]string
	prevOthers []string
	nextServer int
	MyName     string
}

const max = math.MaxUint32

// Update ...
func (sDHT *SimplisticDHT) Update(otherServers []string) {
	// compare otherservers to prevOthers to see if we want to go through this
	// entire process
	if sDHT.compareArrays(otherServers) {
		// log.Print("no change, returning")
		return
	}
	var otherServersHashes []int
	for _, e := range otherServers {
		h := simpleASCIIHash(e, max)
		sDHT.DataMap[h] = e
		otherServersHashes = append(otherServersHashes, h)
	}
	// log.Print("datamap", sDHT.DataMap)
	sDHT.assignSubsequents(otherServersHashes)

	sDHT.prevOthers = otherServers[:]
}

// assign subsequent server to nextServer pointer for this server
func (sDHT *SimplisticDHT) assignSubsequents(otherServersHashes []int) {
	//add my hash to the list, sort the list and find my idex
	myHash := simpleASCIIHash(sDHT.MyName, max)
	otherServersHashes = append(otherServersHashes, myHash)
	sort.Ints(otherServersHashes)
	myIndex := -1
	for i, e := range otherServersHashes {
		if e == myHash {
			myIndex = i
			break
		}
	}

	//I am the last element, need to point at first
	if myIndex == len(otherServersHashes) {
		sDHT.nextServer = otherServersHashes[0]
	} else {
		sDHT.nextServer = otherServersHashes[myIndex+1]
	}
}

// WHO ...
func (sDHT *SimplisticDHT) Who(query string) string {
	queryHash := simpleASCIIHash(query, max)
	log.Printf("Looking for %v which has hash %v \n", query, queryHash)
	return "Bopo"
}

func (sDHT *SimplisticDHT) compareArrays(otherServers []string) bool {
	if len(sDHT.prevOthers) == 0 {
		return false
	}
	for i, e := range otherServers {
		if e != sDHT.prevOthers[i] {
			return false
		}
	}
	return true
}
