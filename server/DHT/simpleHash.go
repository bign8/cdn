package DHT

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
