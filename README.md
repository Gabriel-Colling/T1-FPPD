# T1 — Problemas Clássicos de Concorrência (FPPD)

Trabalho em **Go** com dois problemas de programação concorrente:

1. **Jantar dos Filósofos**: versão base com deadlock, hierarquia de recursos e limitação de `N-1` filósofos.
2. **Produtor/Consumidor com Buffer Limitado**: versão com channel e versão com semáforos.

## Organização do projeto

```text
T1-FPPD-organizado/
├── go.mod
├── README.md
├── filosofos/
│   ├── base/
│   │   └── versao_base.go
│   ├── hierarquia/
│   │   └── versao_hierarquia.go
│   └── limite/
│       └── versao_limite.go
└── produtor_consumidor/
    ├── channels/
    │   └── channels_pc.go
    └── semaforos/
        └── sem_pc.go
```

Há **apenas um `go.mod`**, na raiz: todos os programas fazem parte do mesmo módulo Go. Cada variante fica em uma pasta própria para não haver conflito entre funções `main()` e outras declarações do pacote `main`.

## Pré-requisitos

- Go **1.20 ou superior** instalado (verifique com `go version`).
- Terminal PowerShell, Prompt de Comando ou Bash.
- Não são utilizadas bibliotecas externas.

**Todos os comandos abaixo devem ser executados na pasta raiz `T1-FPPD/`**, isto é, a pasta onde está o `go.mod`. Abra essa pasta no VS Code ou use `cd` no terminal para acessá-la.

## Problema 1 — Jantar dos Filósofos

Cada programa recebe dois parâmetros de linha de comando:

- `-n`: quantidade de filósofos, com `N >= 2` (padrão: `5`);
- `-r`: quantidade de refeições por filósofo, com `R >= 1` (padrão: `3`).

### Versão base — demonstração do deadlock

```bash
go run ./filosofos/base -n 5 -r 3
```

Nesta **versão original**, todos pegam primeiro o garfo esquerdo; uma barreira faz com que tentem pegar o direito depois de todos segurarem o esquerdo. **O programa entra em deadlock propositalmente e não encerra normalmente.** Para interromper, use `Ctrl+C`. É necessário adicionar um mecanismo de cancelamento para cumprir o requisito de encerramento ordenado do enunciado.

### Hierarquia de recursos — sem deadlock

```bash
go run ./filosofos/hierarquia -n 5 -r 3
```

Cada filósofo pega primeiro o garfo de menor índice e depois o de maior índice, evitando espera circular.

### Limitação de `N-1` filósofos — sem deadlock

```bash
go run ./filosofos/limite -n 5 -r 3
```

Um channel de permissões permite que, no máximo, `N-1` filósofos disputem os garfos simultaneamente.

## Problema 2 — Produtor/Consumidor

Nas duas versões disponíveis, o programa usa **P = 5 produtores, C = 5 consumidores e 200 itens por produtor** (1.000 itens no total). Esses valores estão definidos como constantes em cada `main()`; não são parâmetros da linha de comando.

Cada versão executa automaticamente os testes de capacidade **K = 1, 10 e 100** e exibe tempo total, throughput, ocupação média estimada e itens consumidos por consumidor.

### Buffer implementado com channel

```bash
go run ./produtor_consumidor/channels
```

O channel com capacidade `K` faz o produtor esperar quando o buffer está cheio e o consumidor esperar quando está vazio. O buffer é fechado depois que todos os produtores terminam, permitindo que os consumidores finalizem após a drenagem.

### Buffer implementado com semáforos

```bash
go run ./produtor_consumidor/semaforos
```

São usados os semáforos de contagem `notFull` e `notEmpty`, além de mutexes para proteger as posições de escrita e leitura. O encerramento desta versão usa um indicador de conclusão e um timeout de 150 ms.

**Observação:** nesta versão, `len(notEmpty)` é utilizado como estimativa da ocupação, e os arquivos originais ainda não imprimem uma comparação explícita entre totais produzidos e consumidos. Confirme e ajuste essas partes antes da entrega final, conforme o enunciado.

## Verificar condições de corrida (data race)

Execute o detector de corrida do Go nas versões que encerram normalmente:

```bash
go run -race ./filosofos/hierarquia -n 5 -r 3
go run -race ./filosofos/limite -n 5 -r 3
go run -race ./produtor_consumidor/channels
go run -race ./produtor_consumidor/semaforos
```

No Windows, a opção `-race` exige **CGO habilitado e um compilador C compatível** (como GCC). Caso apareça `go: -race requires cgo` ou `gcc not found`, configure o ambiente antes de tentar novamente. O uso do `-race` monitora condições de corrida na execução; não comprova ausência de deadlock. A versão base, como está neste conjunto de arquivos, fica bloqueada intencionalmente.

Para verificar a compilação de todos os pacotes sem executar as simulações:

```bash
go build ./...
```


