package library

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomMoney generates a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomOwner returns a random string of length 8
func RandomOwner() string {
	return RandomString(8)
}

// RandomBalance returns a random bank account balance bigint. Balance is represented in cents
func RandomBalance() int64 {
	return RandomInt(1, 5000000)
}

// RandomCurrency returns a random currency string
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

// // RandomFloatString returns a random string of a float in range from min to max, with 2 decimal numbers
// func randomFloatString(min, max float64) string {
// 	f := min + rand.Float64()*(max-min)
// 	return fmt.Sprintf("%.2f", f)
// }

// RandomString returns a random string of length n
func RandomString(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	k := len(alphabet)

	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}
