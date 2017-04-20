package DHT

type DHT interface {
	Who(string) string
	Update([]string)
}
