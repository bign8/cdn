package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func sum(vals []interface{}) interface{} {
	switch vals[0].(type) {
	case int:
		var acc int
		for _, val := range vals {
			acc += val.(int)
		}
		return acc
	case float64:
		var acc float64
		for _, val := range vals {
			acc += val.(float64)
		}
		return acc
	default:
		fmt.Printf("??? - %T\n", vals[0])
		panic("unknown type")
	}
}

func avg(vals []interface{}) interface{} {
	res := sum(vals)
	switch res.(type) {
	case int:
		return res.(int) / len(vals)
	case float64:
		return res.(float64) / float64(len(vals))
	default:
		fmt.Printf("???? - %T\n", res)
		panic("unknown tyhpe")
	}
}

func main() {
	fh, err := os.Open("data.json")
	check(err)
	decoder := json.NewDecoder(fh)

	for err == nil {
		var obj map[string]interface{}
		err = decoder.Decode(&obj)
		if err != nil {
			break
		}

		// Lose the unique server prefixes
		lists := make(map[string][]interface{}, len(obj))
		for key, value := range obj {
			parts := strings.Split(key, ".")
			if len(parts) > 2 {
				parts = append([]string{parts[0]}, parts[2:]...)
			}
			key = strings.Join(parts, ".")
			lists[key] = append(lists[key], value)
		}

		// aggregate data as desired
		obj = make(map[string]interface{}, len(lists))
		for key, values := range lists {
			last := key[strings.LastIndex(key, ".")+1:]
			last = last[strings.LastIndex(last, "-")+1:]

			// summations
			switch last {
			case "bad":
				fallthrough
			case "img":
				fallthrough
			case "count":
				fallthrough
			case "cacheSize":
				fallthrough
			case "neighbor_list":
				fallthrough
			case "s2s_calls":
				fallthrough
			case "neighbor_miss":
				fallthrough
			case "neighbor_hit":
				obj[key] = sum(values)

				// Average
			case "minute":
				fallthrough
			case "percentile":
				fallthrough
			case "rate":
				fallthrough
			case "max":
				fallthrough
			case "min":
				fallthrough
			case "dev":
				fallthrough
			case "mean":
				obj[key] = avg(values)
				if key == "client.render.max" {
					obj["count"] = len(values)
				}

				// identity
			case "uptime":
				obj[key] = values[0]

			default:
				fmt.Println("Unknown kehy", last)
			}
		}

		// TODO: generate a CSV for plotting
		fmt.Println("----------------------")
		fmt.Println("count", obj["count"])
		fmt.Println("neigh_hit", obj["server.neighbor_hit"])
		fmt.Println("neigh_miss", obj["server.neighbor_miss"])
		fmt.Println("client.render.(min,mean,99,max)", obj["client.render.min"], obj["client.render.mean"], obj["client.render.99-percentile"], obj["client.render.max"])
		// fmt.Printf("?? %#v\n", obj)
	}
}
