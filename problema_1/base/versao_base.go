package main

import (
	"flag"
	"fmt"
	"sync"
)

func filosofo(
	id int,
	R int,
	garfos []chan struct{},
	barreira *sync.WaitGroup,
	wg *sync.WaitGroup) {
	defer wg.Done()

	N := len(garfos)

	esquerdo := garfos[id]
	direito := garfos[(id+1)%N]

	for i := 0; i < R; i++ {
		fmt.Printf("Filósofo %d está pensando", id)

		// pega o garfo esquerdo
		<-esquerdo
		fmt.Printf("Filósofo %d pegou o garfo esquerdo", id)

		// garante que todos peguem o esquerdo
		if i == 0 {
			barreira.Done()
			barreira.Wait()
		}

		// tenta pegar o garfo direito
		fmt.Printf("Filósofo %d tenta pegar o direito\n", id)
		<-direito

		fmt.Printf("Filósofo %d está comendo\n", id)

		// devolve os garfos
		esquerdo <- struct{}{}
		direito <- struct{}{}
	}
}

func main() {
	N := flag.Int("n", 5, "Número de filósofos")
	R := flag.Int("r", 3, "Número de refeições")
	flag.Parse()

	if *N < 2 || *R < 1 {
		fmt.Println("N deve ser >= 2 e R >= 1")
		return
	}

	garfos := make([]chan struct{}, *N)

	for i := 0; i < *N; i++ {
		garfos[i] = make(chan struct{}, 1)
		garfos[i] <- struct{}{}
	}

	var barreira sync.WaitGroup
	barreira.Add(*N)

	var wg sync.WaitGroup

	for i := 0; i < *N; i++ {
		wg.Add(1)
		go filosofo(i, *R, garfos, &barreira, &wg)
	}

	wg.Wait()
}
