package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

func filosofo(
	id int,
	R int,
	garfos []chan struct{},
	permissao chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	N := len(garfos)
	mesa := permissao
	esquerdo := garfos[id]
	direito := garfos[(id+1)%N]

	for i := 0; i < R; i++ {
		fmt.Printf("Filósofo %d está pensando\n", id)
		time.Sleep(200 * time.Millisecond)
		mesa <- struct{}{}

		// Pega o garfo esquerdo
		<-esquerdo
		fmt.Printf("Filósofo %d pegou o garfo esquerdo\n", id)

		// tenta pegar o garfo direito
		<-direito
		fmt.Printf("Filósofo %d tenta pegar o direito\n", id)

		fmt.Printf("Filósofo %d está comendo\n", id)
		time.Sleep(100 * time.Millisecond)
		// devolve os garfos
		esquerdo <- struct{}{}
		direito <- struct{}{}

		<-mesa
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
	permissao := make(chan struct{}, *N-1)
	for i := 0; i < *N; i++ {
		garfos[i] = make(chan struct{}, 1)
		garfos[i] <- struct{}{}
	}

	var barreira sync.WaitGroup
	barreira.Add(*N)

	var wg sync.WaitGroup

	for i := 0; i < *N; i++ {
		wg.Add(1)
		go filosofo(i, *R, garfos, permissao, &wg)
	}

	wg.Wait()
}
