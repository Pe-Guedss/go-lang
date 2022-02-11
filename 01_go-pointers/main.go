package main

import "fmt"

func swap(num1 *int, num2 *int) (int, int) {
	aux := int(0)
	aux = *num1
	*num1 = *num2
	*num2 = aux

	return *num1, *num2
}

func main() {
	x := int(5)
	y := int(21)
	x, y = swap(&x, &y)
	fmt.Println(x, y)
}
