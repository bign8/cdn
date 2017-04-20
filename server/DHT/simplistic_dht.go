package DHT

import (
	"hash/crc32"
	"log"
	"math"
	"sort"
)

//SimplisticDHT ... Simplistic hashtable, assumes nodes (cdns) hold data in linear order, simply decrement close to query hash to find owner
type SimplisticDHT struct {
	DataMap    map[int]*Pair
	prevOthers []string
	nextServer int
	MyName     string
	MyHash     int
	lastServer bool
}

type Pair struct {
	name      string
	subServer int //server next in list
}

const max = math.MaxUint32

// Update ...
func (sDHT *SimplisticDHT) Update(otherServers []string) {
	// compare otherservers to prevOthers to see if we want to go through this
	// entire process

	otherServers = append(otherServers, sDHT.MyName)
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
		h := hash(e, max)
		sDHT.DataMap[h] = &Pair{name: e, subServer: -1}
		otherServersHashes = append(otherServersHashes, h)
	}
	log.Print("DHT's datamap", sDHT.DataMap)

	sDHT.assignSubsequents(otherServersHashes)
	log.Print(sDHT.MyHash, "->", sDHT.DataMap[sDHT.MyHash].subServer)

	sDHT.prevOthers = otherServers[:]
}

func (sDHT *SimplisticDHT) assignSubsequents(otherServersHashes []int) {
	sDHT.MyHash = hash(sDHT.MyName, max)
	sort.Ints(otherServersHashes)
	for i, e := range otherServersHashes {
		if i == len(otherServersHashes)-1 {
			sDHT.DataMap[e].subServer = otherServersHashes[0]
		} else {
			sDHT.DataMap[e].subServer = otherServersHashes[i+1]
		}
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
func (sDHT *SimplisticDHT) Who(query string) string {
	queryHash := hash(query, max)
	log.Printf("Looking for %v in DHT which has hash %v \n", query, queryHash)
	maxK := -1

	for k := range sDHT.DataMap {
		if queryHash > k && queryHash < sDHT.DataMap[k].subServer {
			return sDHT.DataMap[k].name
		} else {
			if k > maxK {
				maxK = k
			}
		}
	}
	return sDHT.DataMap[maxK].name
}

func hash(input string, capacity uint32) int {
	return int(crc32.ChecksumIEEE([]byte(input)) % capacity)
}
