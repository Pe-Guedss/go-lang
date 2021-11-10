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

func (c *Circle) area() float64 {
	return math.Pi * math.Pow(c.radius, 2)
}

func (c Circle) print(num int) {
	fmt.Printf("\nNúmero do círculo: %d\n", num)
	fmt.Printf("Raio: %f\n", c.radius)
	fmt.Printf("X: %f  --  Y: %f\n", c.coord_x, c.coord_y)
	fmt.Printf("Área: %f\n\n", c.area_cir)
	sum_areas += c.area_cir
}

//Retângulos e quadrados
type Rectangle struct {
	coord_x   float32
	coord_y   float32
	len       float64
	wid       float64
	area_rect float64
}

func (r Rectangle) construtor(num int) {
	for true {
		r.init()
		fmt.Printf("len: %f  || wid: %f", r.len, r.wid)
		if !(r.len == 0 || r.wid == 0) {
			break
		}
	}
	r.print(num)
}

func (r *Rectangle) init() {
	r.coord_x = rand.Float32() * float32(rand.Intn(100))
	r.coord_y = rand.Float32() * float32(rand.Intn(100))
	r.len = rand.Float64() * float64(rand.Intn(10))
	r.wid = rand.Float64() * float64(rand.Intn(10))
	r.area_rect = r.area()
}

func (r *Rectangle) area() float64 {
	return r.len * r.wid
}

func (r Rectangle) print(num int) {
	fmt.Printf("\nNúmero do retângulo: %d\n", num)
	fmt.Printf("Comprimento: %f  --  Largura: %f\n", r.len, r.wid)
	fmt.Printf("X inicial: %f  --  Y inicial: %f\n", r.coord_x, r.coord_y)
	fmt.Printf("Área: %f\n\n", r.area())
	sum_areas += r.area()
}

var sum_areas float64

func main() {
	sum_areas = 0
	num := int(0)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("\nOcorreu um erro (panic!) terminal!\n")
		}
		if num < 0 {
			fmt.Printf("\nO número não pode ser negativo.\n")
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

	fmt.Printf("A soma total das áreas é: %f", sum_areas)
}
