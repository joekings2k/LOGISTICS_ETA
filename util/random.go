package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)
const alphabet  = "abcdefghijklmnopqrstuvwxyz"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomInt(min,max int64)int64{
	return min + seededRand.Int63n(max-min+1) //interger btween min and max
}

func RandomString(n int)string {
	var sb strings.Builder
	k := len(alphabet)

	for i:= 0; i <n ; i++ {
		c:=alphabet[seededRand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomEmail()string {
	return fmt.Sprintf("%s@gmail.com", RandomString(6))}

func RandomRole() string {
	roles := []string{"admin", "driver", "customer"}
	n := len(roles)
	return roles[seededRand.Intn(n)]
}