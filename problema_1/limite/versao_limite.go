package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

type Metricas struct {
	Refeicoes       int
	EsperaGarfos    time.Duration
	EsperaPermissao time.Duration
}

func filosofo(
	id int,
	R int,
	garfos []chan struct{},
	permissao chan struct{},
	metricas []Metricas,
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

		inicioPermissao := time.Now()
		mesa <- struct{}{}

		metricas[id].EsperaPermissao += time.Since(inicioPermissao)

		inicioGarfos := time.Now()

		// Pega o garfo esquerdo
		<-esquerdo
		fmt.Printf("Filósofo %d pegou o garfo esquerdo\n", id)

		// tenta pegar o garfo direito
		<-direito
		fmt.Printf("Filósofo %d tenta pegar o direito\n", id)

		metricas[id].EsperaGarfos += time.Since(inicioGarfos)

		fmt.Printf("Filósofo %d está comendo\n", id)
		time.Sleep(200 * time.Millisecond)

		metricas[id].Refeicoes++

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

	metricas := make([]Metricas, *N)

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

	var wg sync.WaitGroup

	for i := 0; i < *N; i++ {
		wg.Add(1)
		go filosofo(i, *R, garfos, permissao, metricas, &wg)
	}

	wg.Wait()

	fmt.Println("\n=== RESULTADOS DO LIMITE N-1 ===")

	for i, m := range metricas {
		mediaGarfos := m.EsperaGarfos.Seconds() * 1000 / float64(m.Refeicoes)
		mediaPermissao := m.EsperaPermissao.Seconds() * 1000 / float64(m.Refeicoes)

		fmt.Printf(
			"Filósofo %d: %d refeições | Garfos: %.3f ms | Permissão: %.3f ms | Total: %.3f ms\n",
			i, m.Refeicoes, mediaGarfos, mediaPermissao,
			mediaGarfos+mediaPermissao,
		)
	}
}
