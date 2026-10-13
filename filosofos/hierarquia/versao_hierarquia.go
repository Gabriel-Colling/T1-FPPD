package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

type Metricas struct {
	Refeicoes   int
	TempoEspera time.Duration
}

func filosofoHierarquico(
	id int,
	R int,
	garfos []chan struct{},
	metricas []Metricas,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	N := len(garfos)

	esquerdo := garfos[id]
	direito := garfos[(id+1)%N]

	for i := 0; i < R; i++ {
		fmt.Printf("Filósofo %d está pensando\n", id)

		inicio := time.Now()

		if id < (id+1)%N {
			<-esquerdo
			fmt.Printf("Filósofo %d pegou o garfo esquerdo\n", id)
			<-direito
			fmt.Printf("Filósofo %d pegou o garfo direito\n", id)
		} else {
			<-direito
			fmt.Printf("Filósofo %d pegou o garfo direito\n", id)
			<-esquerdo
			fmt.Printf("Filósofo %d pegou o garfo esquerdo\n", id)
		}

		metricas[id].TempoEspera += time.Since(inicio)
		fmt.Printf("Filósofo %d está comendo\n", id)

		metricas[id].Refeicoes++

		// Devolve os garfos
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

	metricas := make([]Metricas, *N)

	var wg sync.WaitGroup
	for i := 0; i < *N; i++ {
		wg.Add(1)
		go filosofoHierarquico(i, *R, garfos, metricas, &wg)
	}

	wg.Wait()
	fmt.Println("\n=== RESULTADOS DA HIERARQUIA ===")

	for i, m := range metricas {
		media := m.TempoEspera.Seconds() * 1000 / float64(m.Refeicoes)

		fmt.Printf(
			"Filósofo %d: %d refeições | Espera média: %.3f ms\n",
			i, m.Refeicoes, media,
		)
	}
}
