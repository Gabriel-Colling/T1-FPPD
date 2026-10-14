
# T1 — Problemas Clássicos de Concorrência

**Disciplina:** Fundamentos de Processamento Paralelo e Distribuído (FPPD)

**Linguagem:** Go

## 1. Descrição do Projeto

Este trabalho implementa dois problemas clássicos de concorrência utilizando goroutines, channels e mecanismos de sincronização da linguagem Go.

Os problemas desenvolvidos são:

- **Problema 1 — Jantar dos Filósofos:** implementação de uma versão base com deadlock e duas estratégias de prevenção: hierarquia de recursos e limitação de N-1 filósofos.
- **Problema 2 — Produtor/Consumidor:** implementação de um buffer limitado utilizando channels e outra utilizando semáforos de contagem.

Cada versão possui um programa independente, permitindo executar e comparar as diferentes estratégias.

## 2. Estrutura do Projeto

```text
T1-FPPD/
│
├── go.mod
├── README.md
│
├── filosofos/
│   ├── base/
│   │   └── versao_base.go
│   ├── hierarquia/
│   │   └── versao_hierarquia.go
│   └── limite/
│       └── versao_limite.go
│
└── produtor-consumidor/
    ├── PC_channels/
    │   └── channels_pc.go
    └── PC_sem/
        └── sem_pc.go
```

O projeto utiliza um único arquivo `go.mod` na raiz, compartilhado pelos dois problemas. Todas as dependências utilizadas pertencem à biblioteca padrão do Go.

## 3. Pré-requisitos

Para executar os programas, é necessário:

- Go 1.20 ou superior.
- Terminal PowerShell, Prompt de Comando ou Bash.
- GCC e CGO habilitado para executar os testes com `-race` no Windows.

Para verificar a instalação do Go:

```bash
go version
```

**Importante:** todos os comandos devem ser executados na pasta principal do projeto, onde está localizado o arquivo `go.mod`.

---

## 4. Problema 1 — Jantar dos Filósofos

O problema representa N filósofos sentados em uma mesa circular, compartilhando N garfos.

Cada filósofo é representado por uma goroutine e cada garfo por um channel com capacidade 1.

Os programas recebem dois parâmetros:

- `-n`: quantidade de filósofos (N >= 2). Padrão: 5.
- `-r`: quantidade de refeições por filósofo (R >= 1). Padrão: 3.

### 4.1 Versão Base — Deadlock

**Execução:**

```bash
go run ./filosofos/base -n 5 -r 3
```

Todos os filósofos tentam adquirir primeiro o garfo esquerdo e depois o direito.

Uma barreira de sincronização garante que todos adquiram o garfo esquerdo antes de tentar adquirir o direito, provocando intencionalmente um deadlock.

Após demonstrar o bloqueio, o programa utiliza um channel de cancelamento e `select` para interromper as tentativas de aquisição.

As goroutines devolvem os recursos e finalizam, enquanto `sync.WaitGroup` garante o encerramento ordenado.

Como o deadlock é provocado antes da primeira refeição, a média de espera por refeição é apresentada como não disponível.

### 4.2 Hierarquia de Recursos

**Execução:**

```bash
go run ./filosofos/hierarquia -n 5 -r 3
```

Os filósofos adquirem primeiro o garfo com menor índice e depois o de maior índice.

Essa ordem global elimina a condição de espera circular de Coffman, evitando deadlock.

Cada filósofo executa R refeições, devolvendo os garfos após cada uma. Ao finalizar, `sync.WaitGroup` aguarda todas as goroutines terminarem.

O programa apresenta o número de refeições e o tempo médio de espera pelos garfos de cada filósofo.

### 4.3 Limitação de N-1 Filósofos

**Execução:**

```bash
go run ./filosofos/limite -n 5 -r 3
```

A estratégia utiliza um channel de permissões com capacidade N-1, funcionando como um semáforo.

Cada filósofo precisa obter uma permissão antes de disputar os garfos. Após comer e devolver os dois garfos, libera sua permissão.

Essa limitação impede que todos os filósofos formem simultaneamente um ciclo completo de espera.

O programa apresenta a quantidade de refeições e os tempos médios de espera pela permissão, pelos garfos e o tempo total.

### 4.4 Métricas do Problema 1

As versões registram:

- Número de refeições concluídas por filósofo.
- Tempo de espera para adquirir os garfos.
- Tempo médio de espera por refeição.
- Tempo de espera pela permissão de entrada, na estratégia N-1.

Nas versões hierárquica e N-1, são utilizados intervalos simulados de 200 ms para pensar e 200 ms para comer.

Os resultados permitem comparar a distribuição dos tempos de espera e analisar a **fairness** das estratégias.

---

## 5. Problema 2 — Produtor/Consumidor com Buffer Limitado

O problema simula produtores inserindo itens em um buffer compartilhado e consumidores retirando esses itens.

As duas versões utilizam os seguintes valores:

- **P = 5** produtores.
- **C = 5** consumidores.
- **200 itens** produzidos por produtor.
- **1.000 itens** no total por execução.
- Capacidades **K = 1, 10 e 100**.

Os valores P, C e a quantidade de itens estão definidos como constantes nos programas. Cada execução testa automaticamente as três capacidades K.

### 5.1 Versão com Channels

**Execução:**

```bash
go run ./produtor-consumidor/PC_channels
```

O buffer é implementado utilizando um channel com capacidade K.

Quando o buffer está cheio, os produtores ficam bloqueados aguardando espaço. Quando está vazio, os consumidores aguardam novos itens.

Após o encerramento dos produtores, o channel é fechado. Os consumidores continuam retirando os itens restantes e finalizam quando o buffer é drenado.

O programa também possui um `select` com `time.After(150 * time.Millisecond)` para tratar períodos de espera por itens.

### 5.2 Versão com Semáforos

**Execução:**

```bash
go run ./produtor-consumidor/PC_sem
```

O buffer é representado por um array circular compartilhado de capacidade K.

A sincronização utiliza dois semáforos de contagem:

- `notFull`: controla a disponibilidade de posições para escrita.
- `notEmpty`: controla a disponibilidade de itens para leitura.

São utilizados mutexes para proteger as operações de escrita e leitura, impedindo acessos simultâneos incorretos aos respectivos índices.

Após o término dos produtores, um indicador de conclusão é atualizado. Os consumidores utilizam `select` com timeout para verificar se a produção terminou e se não existem mais itens disponíveis, encerrando suas goroutines.

### 5.3 Métricas do Problema 2

As duas versões executam testes para K = 1, 10 e 100, coletando:

- Tempo total de execução.
- Throughput, medido em itens por segundo.
- Ocupação média estimada do buffer.
- Quantidade de itens consumidos por consumidor.
- Total de itens produzidos e consumidos.

O throughput é calculado dividindo o número de itens produzidos pelo tempo total da execução.

A ocupação média é estimada por amostragens periódicas de aproximadamente 1 ms.

Na versão com semáforos, utiliza-se `len(notEmpty)` como aproximação da quantidade de itens disponíveis.

Ao final, ambas as implementações verificam se o total produzido corresponde ao total consumido.

Essas métricas permitem analisar o efeito da capacidade K no desempenho e no desacoplamento entre produtores e consumidores.

---

## 6. Verificação de Data Race

Para verificar possíveis condições de corrida, utilize a opção `-race` do Go.

### Problema 1 — Jantar dos Filósofos

**Versão Base:**

```bash
go run -race ./filosofos/base -n 5 -r 3
```

**Hierarquia de Recursos:**

```bash
go run -race ./filosofos/hierarquia -n 5 -r 3
```

**Limitação N-1:**

```bash
go run -race ./filosofos/limite -n 5 -r 3
```

### Problema 2 — Produtor/Consumidor

**Versão com Channels:**

```bash
go run -race ./produtor-consumidor/PC_channels
```

**Versão com Semáforos:**

```bash
go run -race ./produtor-consumidor/PC_sem
```

O detector analisa os acessos concorrentes à memória durante a execução e informa possíveis condições de corrida por meio de mensagens como:

```text
WARNING: DATA RACE
```

A ausência de alertas em uma execução não garante formalmente que não existam condições de corrida em outros escalonamentos.

**Observação para Windows:** o detector `-race` depende de CGO habilitado e de um compilador C compatível, como GCC.

---

## 7. Compilação e Verificação

Para verificar se todos os programas compilam corretamente:

```bash
go build ./...
```

Para executar a análise estática do Go:

```bash
go vet ./...
```

Para conferir as dependências do projeto:

```bash
go mod tidy
```

Não é necessário instalar bibliotecas externas, pois o projeto utiliza apenas pacotes da biblioteca padrão do Go.

---

## 8. Observações sobre a Implementação

- A versão base do Jantar dos Filósofos provoca deadlock de forma controlada e encerra antecipadamente após o cancelamento.
- As versões hierárquica e N-1 realizam as R refeições antes de encerrar.
- A versão com channels do Produtor/Consumidor possui um caso de timeout de 150 ms. Caso esse ramo ainda esteja vazio, é necessário adicionar uma ação alternativa para atender integralmente ao enunciado.
- A ocupação média da versão com semáforos é estimada pela quantidade de tokens disponíveis em `notEmpty`, não representando necessariamente a ocupação física exata em todos os instantes.
- As métricas podem variar entre execuções devido ao escalonamento das goroutines.

---

