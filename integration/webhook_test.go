package integration

import (
	"fmt"
	"testing"
)

func Test_setup(t *testing.T) {
	fmt.Println("SALUT")
	beforeTest(t)
	fmt.Println("SALUT2")
}
