package main

import "fmt"

func odd_or_even(num int) (int, bool) {
	if num < 0 {
		return 0, false
	} else if num%2 == 0 {
		return num / 2, true
	}
	return num / 2, false
}

func bigger(args ...int) int {
	maior := args[0]
	for i := range args {
		if args[i] > maior {
			maior = args[i]
		}
	}
	return maior
}

func makeOddGenerator() func() uint16 {
	i := uint16(1)
	return func() (ret uint16) {
		ret = i
		i += 2
		return
	}
}

func main() {
	num := 0
	fmt.Print("Digite um número para testar paridade: ")
	fmt.Scanln(&num)
	res, parity := odd_or_even(num)
	fmt.Println(res, parity)
	fmt.Println("\n-----------------------------------------\n ")
	lista := []int{1, 2, 3, 50, 48, 477, 10000, 514, 89, 45, 124, 63145, 6841, 15245}
	fmt.Print("O maior número de: ")
	for i := 0; i < len(lista); i++ {
		fmt.Printf("%d, ", lista[i])
	}
	fmt.Printf("\nÉ o número: %d\n", bigger(lista...))
	fmt.Println("\n-----------------------------------------\n ")
	c := "a"
	nextOdd := makeOddGenerator()
	fmt.Println("Digite \"s\" para parar a execução. Digite qualquer outra letra para continuar a gerar números ímpares!\n ")
	for c != "s" {
		fmt.Printf("%d ", nextOdd())
		fmt.Scan(&c)
	}
}
