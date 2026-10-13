package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Resultado struct {
	K                    int
	TempoTotal           time.Duration
	Throughput           float64
	OcupacaoMedia        float64
	ConsumoPorConsumidor []int32
}

func main() {
	const P = 5
	const C = 5
	const ItensPorP = 200

	capacidades := []int{1, 10, 100}
	fmt.Println("=== INICIANDO BENCHMARK (CHANNELS) ===")
	for _, K := range capacidades {
		r := executarTesteChannel(K, P, C, ItensPorP)
		imprimirRelatorio(r)
	}
}

func executarTesteChannel(K, P, C, ItensPorP int) Resultado {
	totalItens := P * ItensPorP
	consumoPorConsumidor := make([]int32, C)

	// --- SETUP DO MONITOR (Coleta de Dados) ---
	buffer := make(chan int, K) // O Buffer precisa ser declarado aqui para o monitor poder ler
	doneMonitor := make(chan struct{})
	monitorResult := make(chan float64)

	go func() {
		var soma, amostras int
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-doneMonitor:
				media := 0.0
				if amostras > 0 {
					media = float64(soma) / float64(amostras)
				}
				monitorResult <- media
				return
			case <-ticker.C:
				soma += len(buffer) // Verifica o tamanho atual do buffer
				amostras++
			}
		}
	}()
	// ------------------------------------------

	inicio := time.Now()

	// =========================================================================
	// ⬇️ INÍCIO DA LÓGICA ORIGINAL (PRODUTOR/CONSUMIDOR COM CHANNELS) ⬇️
	// =========================================================================
	var wgProdutores sync.WaitGroup
	var wgConsumidores sync.WaitGroup

	// Produtores
	for i := 0; i < P; i++ {
		wgProdutores.Add(1)
		go func(prodID int) {
			defer wgProdutores.Done()
			for j := 0; j < ItensPorP; j++ {
				buffer <- 1                      // Bloqueia se cheio
				time.Sleep(1 * time.Millisecond) // Simula produção
			}
		}(i)
	}

	// Consumidores
	for i := 0; i < C; i++ {
		wgConsumidores.Add(1)
		go func(consID int) {
			defer wgConsumidores.Done()
			for {
				select {
				case _, ok := <-buffer:
					if !ok {
						return // Drenado e fechado
					}
					// Incrementa contador específico deste consumidor para métrica
					atomic.AddInt32(&consumoPorConsumidor[consID], 1)
					time.Sleep(2 * time.Millisecond) // Simula consumo (mais lento que a produção)

				case <-time.After(150 * time.Millisecond):
					// Timeout
				}
			}
		}(i)
	}

	wgProdutores.Wait()
	close(buffer)
	wgConsumidores.Wait()

	totalConsumido := 0

	for _, quantidade := range consumoPorConsumidor {
		totalConsumido += int(quantidade)
	}

	fmt.Printf(
		"Produzidos: %d | Consumidos: %d\n",
		totalItens, totalConsumido,
	)

	if totalConsumido != totalItens {
		panic("Erro: quantidade produzida diferente da consumida")
	}

	// =========================================================================
	// ⬆️ FIM DA LÓGICA ORIGINAL ⬆️
	// =========================================================================

	fim := time.Now()
	tempoTotal := fim.Sub(inicio)

	// Para o monitor e coleta o resultado da ocupação
	close(doneMonitor)
	ocupacaoMedia := <-monitorResult

	return Resultado{
		K:                    K,
		TempoTotal:           tempoTotal,
		Throughput:           float64(totalItens) / tempoTotal.Seconds(),
		OcupacaoMedia:        ocupacaoMedia,
		ConsumoPorConsumidor: consumoPorConsumidor,
	}
}

func imprimirRelatorio(r Resultado) {
	fmt.Printf("\n[ K = %-3d ] --------------------------------------------\n", r.K)
	fmt.Printf("Tempo Total: %v\n", r.TempoTotal)
	fmt.Printf("Throughput:  %.2f itens/seg\n", r.Throughput)
	fmt.Printf("Ocup. Média: %.2f itens (%.2f%% da capacidade)\n", r.OcupacaoMedia, (r.OcupacaoMedia/float64(r.K))*100)
	fmt.Printf("Distribuição: ")
	for i, qtd := range r.ConsumoPorConsumidor {
		fmt.Printf("C%d:%d ", i+1, int(qtd))
	}
	fmt.Println()
}
