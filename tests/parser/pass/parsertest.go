package tests

import "fmt"

func parserTest() {
	var x int = 10
	fmt.Println(x)

	var y int = 10
	if y > 5 {
		fmt.Println(y)
	}

	var z int = 5
	for z > 0 {
		fmt.Println(z)
		z = z - 1
	}
}
