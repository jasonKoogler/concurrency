package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Philosopher struct {
	Name       string
	FoodAmount int
	LFork      *sync.Mutex
	RFork      *sync.Mutex
	// Eating     bool
}

func NewPhilosopher(name string, foodAmount int) *Philosopher {
	return &Philosopher{
		Name:       name,
		FoodAmount: foodAmount,
	}
}

func (p *Philosopher) Dine(wg *sync.WaitGroup, dineTimes int) {
	defer wg.Done()

	for i := 0; i < dineTimes; i++ {
		p.Think()
		p.Eat()
	}
}

func (p *Philosopher) Eat() {
	p.LFork.Lock()
	p.RFork.Lock()

	fmt.Printf("%s is eating.\n", p.Name)
	time.Sleep(time.Duration(rand.Intn(1000) * int(time.Millisecond)))

	p.RFork.Unlock()
	p.LFork.Unlock()
}

func (p *Philosopher) Think() {
	fmt.Printf("%s is thinking.\n", p.Name)
	time.Sleep(time.Duration(rand.Intn(1000) * int(time.Millisecond)))
}

func main() {
	dineTimes := 3

	forks := make([]*sync.Mutex, 5)
	for i := 0; i < 5; i++ {
		forks[i] = &sync.Mutex{}
	}

	phs := make([]*Philosopher, 5)
	for i := 0; i < 5; i++ {
		phs[i] = &Philosopher{
			Name:  fmt.Sprintf("Philosopher %d", i+1),
			LFork: forks[i],
			RFork: forks[(i+1)%5],
		}
	}

	var wg sync.WaitGroup
	for _, p := range phs {
		wg.Add(1)
		go p.Dine(&wg, dineTimes)
	}

	wg.Wait()

}
