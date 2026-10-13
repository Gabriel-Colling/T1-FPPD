package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Estrutura do Buffer (Parte da lógica original)
type BoundedBuffer struct {
	buf      []int
	k        int
	head     int
	tail     int
	mutHead  sync.Mutex
	mutTail  sync.Mutex
	notEmpty chan struct{}
	notFull  chan struct{}
	done     bool
	mutDone  sync.Mutex
}

func NewBoundedBuffer(capacidade int) *BoundedBuffer {
	bb := &BoundedBuffer{
		buf:      make([]int, capacidade),
		k:        capacidade,
		notEmpty: make(chan struct{}, capacidade),
		notFull:  make(chan struct{}, capacidade),
	}
	for i := 0; i < capacidade; i++ {
		bb.notFull <- struct{}{}
	}
	return bb
}

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
	fmt.Println("=== INICIANDO BENCHMARK (SEMÁFOROS) ===")
	for _, K := range capacidades {
		r := executarTesteSemaforo(K, P, C, ItensPorP)
		imprimirRelatorio(r)
	}
}

func executarTesteSemaforo(K, P, C, ItensPorP int) Resultado {
	totalItens := P * ItensPorP
	consumoPorConsumidor := make([]int32, C)

	bb := NewBoundedBuffer(K)

	// --- SETUP DO MONITOR (Coleta de Dados) ---
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
				// Em semáforos, a ocupação atual é o tamanho do semáforo notEmpty
				soma += len(bb.notEmpty)
				amostras++
			}
		}
	}()
	// ------------------------------------------

	inicio := time.Now()

	// =========================================================================
	// ⬇️ INÍCIO DA LÓGICA ORIGINAL (PRODUTOR/CONSUMIDOR COM SEMÁFOROS) ⬇️
	// =========================================================================
	var wgProdutores sync.WaitGroup
	var wgConsumidores sync.WaitGroup

	// Produtores
	for i := 0; i < P; i++ {
		wgProdutores.Add(1)
		go func(prodID int) {
			defer wgProdutores.Done()
			for j := 0; j < ItensPorP; j++ {
				<-bb.notFull // Wait(notFull)

				bb.mutHead.Lock()
				bb.buf[bb.head] = 1
				bb.head = (bb.head + 1) % bb.k
				bb.mutHead.Unlock()

				bb.notEmpty <- struct{}{} // Signal(notEmpty)

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
				case <-bb.notEmpty: // Wait(notEmpty)
					bb.mutTail.Lock()
					_ = bb.buf[bb.tail] // Simula a leitura
					bb.tail = (bb.tail + 1) % bb.k
					bb.mutTail.Unlock()

					bb.notFull <- struct{}{} // Signal(notFull)

					// Incrementa contador para métrica
					atomic.AddInt32(&consumoPorConsumidor[consID], 1)
					time.Sleep(2 * time.Millisecond) // Simula consumo lento

				case <-time.After(150 * time.Millisecond):
					bb.mutDone.Lock()
					isDone := bb.done
					bb.mutDone.Unlock()

					if isDone && len(bb.notEmpty) == 0 {
						return // Drenado e encerrado
					}
				}
			}
		}(i)
	}

	wgProdutores.Wait()

	bb.mutDone.Lock()
	bb.done = true
	bb.mutDone.Unlock()

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
	// FIM DA LÓGICA ORIGINAL
	// =========================================================================

	fim := time.Now()
	tempoTotal := fim.Sub(inicio)

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
