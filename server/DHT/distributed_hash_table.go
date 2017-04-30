package DHT

// DHT ..
// TODO: figure out what kind of DHT this is!
type DHT interface {
	Who(string) string
	Update([]string)
}
