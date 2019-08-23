package rand

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestDefaultRand(t *testing.T) {
	fmt.Println("1:", rand.Int())
	fmt.Println("2:", rand.Int())
	fmt.Println("3:", rand.Int())
}

func TestString(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(String(12))
	}
}
