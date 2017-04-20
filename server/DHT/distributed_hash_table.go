package DHT

// DHT ..
type DHT interface {
	Who(string) string
	Update([]string)
	NewDHT(string) DHT
}
