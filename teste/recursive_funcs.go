package main

import "fmt"

func fact(num uint) uint {
	if num <= 1 {
		return 1
	}

	return num * fact(num-1)
}

func fib(num uint) uint {
	if num == 1 {
		return 1
	} else if num == 0 {
		return 0
	}
	return fib(num-1) + fib(num-2)
}

func main() {
	num := 0
	fmt.Print("Digite um número para calcular o fatorial: ")
	fmt.Scan(&num)
	fmt.Printf("\n%d", fact(uint(num)))
	fmt.Println("\n-----------------------------------------\n ")
	fmt.Print("Digite um número para descobrir a sequência de fibonacci até ele: ")
	fmt.Scan(&num)
	fmt.Printf("\n%d", fib(uint(num)))
}
