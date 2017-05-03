package DHT

// DHT ..
// TODO (lisa): figure out what kind of DHT this is!
type DHT interface {
	Who(string) string
	Update([]string)
}
