package main

import (
	"fmt"
	"math"
	"math/rand"
)

type Circle struct {
	coord_x float32
	coord_y float32
	radius  float64
}

func (c Circle) area() float64 {
	return math.Pi * math.Pow(c.radius, 2)
}
func (c *Circle) init() {
	c.radius = rand.Float64() * float64(rand.Intn(20))
	c.coord_x = rand.Float32() * float32(rand.Intn(100))
	c.coord_y = rand.Float32() * float32(rand.Intn(100))
}
func (c Circle) print(num int) {
	fmt.Printf("Número do círculo: %d\n", num)
	fmt.Printf("Raio: %f\n", c.radius)
	fmt.Printf("X: %f  --  Y: %f\n", c.coord_x, c.coord_y)
	fmt.Printf("Raio: %f\n\n", c.area())
}

func main() {
	num_cir := int(0)
	fmt.Print("Quantos círculos você quer criar? ")
	fmt.Scan(&num_cir)
	c := make([]Circle, num_cir)
	// var c []*Circle
	for i := 0; i < num_cir; i++ {
		c[i].init()
		if c[i].radius == 0 {
			i--
		}
	}
	for i := 0; i < num_cir; i++ {
		c[i].print(i)
	}
}
