package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

type Metricas struct {
	Refeicoes          int
	EsperaInterrompida time.Duration
}

func filosofo(
	id int,
	R int,
	garfos []chan struct{},
	barreira *sync.WaitGroup,
	cancelar <-chan struct{},
	metricas []Metricas,
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

		fmt.Printf("Filósofo %d tenta pegar o direito\n", id)

		inicioEspera := time.Now()

		select {
		case <-direito:
			metricas[id].EsperaInterrompida = time.Since(inicioEspera)

			// se o garfo ficou disponível após o cancelamento,
			// devolve os dois recursos e encerra.
			direito <- struct{}{}
			esquerdo <- struct{}{}
			return

		case <-cancelar:
			metricas[id].EsperaInterrompida = time.Since(inicioEspera)

			fmt.Printf("Filósofo %d foi interrompido\n", id)

			// libera o garfo esquerdo antes de encerrar.
			esquerdo <- struct{}{}
			return
		}
	}
}

func main() {
	N := flag.Int("n", 5, "Número de filósofos")
	R := flag.Int("r", 3, "Número de refeições")
	flag.Parse()

	cancelar := make(chan struct{})
	metricas := make([]Metricas, *N)

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
		go filosofo(i, *R, garfos, &barreira, cancelar, metricas, &wg)
	}

	// aguarda todos adquirirem o garfo esquerdo.
	barreira.Wait()

	fmt.Println("\nDEADLOCK: todos possuem o garfo esquerdo!")
	fmt.Println("Nenhum consegue adquirir o direito.")

	// Mantém o bloqueio por um pequeno intervalo.
	time.Sleep(200 * time.Millisecond)

	// Cancela a simulação.
	close(cancelar)

	// Aguarda todas as goroutines encerrarem.
	wg.Wait()

	fmt.Println("\n=== RESULTADOS DA VERSÃO BASE ===")

	for i, m := range metricas {
		fmt.Printf(
			"Filósofo %d: %d refeições | Espera interrompida: %.3f ms | Média por refeição: N/D\n",
			i, m.Refeicoes, m.EsperaInterrompida.Seconds()*1000,
		)
	}
}
