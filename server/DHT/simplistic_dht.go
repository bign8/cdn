package DHT

import (
	"log"
	"math"
)

//SimplisticDHT ... Simplistic hashtable, assumes nodes (cdns) hold data in linear order, simply decrement close to query hash to find owner
type SimplisticDHT struct {
	DataMap map[int]string
}

const max = math.MaxUint32

// Update ...
func (sDHT *SimplisticDHT) Update(otherServers []string) {
	for _, e := range otherServers {
		sDHT.DataMap[simpleASCIIHash(e, max)] = e
	}
	log.Print("datamap", sDHT.DataMap)

}

// WHO ...
func (sDHT *SimplisticDHT) Who(query string) string {
	queryHash := simpleASCIIHash(query, max)
	log.Printf("Looking for %v which has hash %v \n", query, queryHash)
	return "Bopo"
}
