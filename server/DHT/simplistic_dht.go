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
	MyHash     int
}

const max = math.MaxUint32

// Update ...
func (sDHT *SimplisticDHT) Update(otherServers []string) {
	// compare otherservers to prevOthers to see if we want to go through this
	// entire process

	if sDHT.compareArrays(otherServers) {
		// log.Print("No change in otherServers, returning without updating DHT")
		return
	} else if len(otherServers) == 0 {
		log.Print("No otherServers yet, returning without updating DHT")
		return
	} else {
		log.Print("Change in otherServers, updating DHT")
	}
	var otherServersHashes []int

	for _, e := range otherServers {
		h := simpleASCIIHash(e, max)
		sDHT.DataMap[h] = e
		otherServersHashes = append(otherServersHashes, h)
	}
	log.Print("DHT's datamap", sDHT.DataMap)

	sDHT.assignSubsequents(otherServersHashes)
	log.Print(sDHT.MyHash, "->", sDHT.nextServer)

	sDHT.prevOthers = otherServers[:]
}

// assign subsequent server to nextServer pointer for this server
func (sDHT *SimplisticDHT) assignSubsequents(otherServersHashes []int) {
	//add my hash to the list, sort the list and find my idex
	sDHT.MyHash = simpleASCIIHash(sDHT.MyName, max)
	otherServersHashes = append(otherServersHashes, sDHT.MyHash)
	sort.Ints(otherServersHashes)
	myIndex := -1
	for i, e := range otherServersHashes {
		if e == sDHT.MyHash {
			myIndex = i
			break
		}
	}

	//I am the last element, need to point at first
	// NOTE: if running manually through redis, the redis server is part of hashes and myIndex should = len(otherServersHashes)
	if myIndex == len(otherServersHashes)-1 {
		sDHT.nextServer = otherServersHashes[0]
	} else {
		sDHT.nextServer = otherServersHashes[myIndex+1]
	}
}

// Iterate through sorted arrays to compare each element
// TODO: Is there really not a library for this?
func (sDHT *SimplisticDHT) compareArrays(otherServers []string) bool {
	if len(sDHT.prevOthers) == 0 {
		return false
	} else if len(sDHT.prevOthers) != len(otherServers) {
		return false
	}
	for i, e := range otherServers {
		if e != sDHT.prevOthers[i] {
			return false
		}
	}
	return true
}

// WHO ...
func (sDHT *SimplisticDHT) Who(query string) (string, bool) {
	queryHash := simpleASCIIHash(query, max)
	log.Printf("Looking for %v which has hash %v \n", query, queryHash)
	maxK := 0
	for k, v := range sDHT.DataMap {
		if k < queryHash {
			// I am the server covering the portion that wraps past 0 on the ring
			if sDHT.nextServer > queryHash {
				return v, true // Return owner and flag that it is owner
			}
		} else {
			if k > maxK {
				maxK = k
			}
		}
	}

	//couldn't find the item, forward it to largest key, as it is likely to own zero and will be the owner
	return sDHT.DataMap[maxK], false // Return largest and forward flag
}
