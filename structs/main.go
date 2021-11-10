package main

import (
	"fmt"
	"math"
	"math/rand"
)

//Círculos
type Circle struct {
	coord_x  float32
	coord_y  float32
	radius   float64
	area_cir float64
}

func (c Circle) construtor(num int) {
	for c.radius == 0 {
		c.init()
	}
	c.print(num)
}

func (c *Circle) init() {
	c.radius = rand.Float64() * float64(rand.Intn(20))
	c.coord_x = rand.Float32() * float32(rand.Intn(100))
	c.coord_y = rand.Float32() * float32(rand.Intn(100))
	c.area_cir = c.area()
}

func (c Circle) area() float64 {
	return math.Pi * math.Pow(c.radius, 2)
}

func (c Circle) print(num int) {
	fmt.Printf("\nNúmero do círculo: %d\n", num)
	fmt.Printf("Raio: %f\n", c.radius)
	fmt.Printf("X: %f  --  Y: %f\n", c.coord_x, c.coord_y)
	fmt.Printf("Área: %f\n\n", c.area_cir)
}

//Retângulos e quadrados
type Rectangle struct {
	coord_x float32
	coord_y float32
	len     float64
	wid     float64
}

func (r Rectangle) construtor(num int) {
	for r.len == 0 || r.wid == 0 {
		r.init()
	}
	r.print(num)
}

func (r Rectangle) init() {
	r.coord_x = rand.Float32() * float32(rand.Intn(100))
	r.coord_y = rand.Float32() * float32(rand.Intn(100))
	r.len = rand.Float64() * float64(rand.Intn(10))
	r.wid = rand.Float64() * float64(rand.Intn(10))
}

func (r Rectangle) area() float64 {
	return r.len * r.wid
}

func (r Rectangle) print(num int) {
	fmt.Printf("\nNúmero do retângulo: %d\n", num)
	fmt.Printf("Comprimento: %f  --  Largura: %f\n", r.len, r.wid)
	fmt.Printf("X inicial: %f  --  Y inicial: %f\n", r.coord_x, r.coord_y)
	fmt.Printf("Área: %f\n\n", r.area())
}

func main() {
	num := int(0)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("\nOcorreu um erro (panic!) terminal!\n")
		}
		if num < 0 {
			fmt.Println("\nO número não pode ser negativo.")
		}
	}()

	fmt.Print("Quantos círculos você quer criar? ")
	fmt.Scan(&num)
	c := make([]Circle, num)
	for i := 0; i < num; i++ {
		c[i].radius = 0
		c[i].construtor(i)
	}

	fmt.Print("Quantos retângulos você quer criar? ")
	fmt.Scan(&num)
	r := make([]Rectangle, num)
	for i := 0; i < num; i++ {
		r[i].len = 0
		r[i].wid = 0
		r[i].construtor(i)
	}
}
