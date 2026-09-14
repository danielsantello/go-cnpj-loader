# go-cnpj-loader

Carregador versionado dos dados públicos de CNPJ disponibilizados pela Receita Federal, desenvolvido em Go para processar grandes volumes de dados no MySQL.

O programa baixa e carrega publicações completas, registra o histórico operacional e cria um novo schema de dados para cada versão. Ele é independente de qualquer API ou aplicação consumidora.

## Como funciona

O fluxo possui duas operações independentes:

1. `download` consulta uma publicação da Receita Federal e baixa seus arquivos ZIP;
2. `load` verifica os arquivos existentes, cria um novo schema versionado e carrega os dados no MySQL.

O comando `load` não inicia downloads. Durante a carga, o conteúdo dos ZIPs é transmitido diretamente ao MySQL, sem extração física dos CSVs.

Cada arquivo é carregado em uma transação própria. Warnings do MySQL são tratados como erros, e schemas incompletos são preservados para diagnóstico.

Uma carga concluída produz um schema versionado, como:

```text
cnpj_2026_08_001
```

O schema permanente `cnpj_loader_control` registra as publicações, seus arquivos, as versões geradas e as migrations aplicadas.

A Receita Federal é considerada a fonte soberana dos dados. O loader realiza validações técnicas, mas não corrige nem valida semanticamente o conteúdo cadastral.

Para conhecer o fluxo completo e as decisões do projeto, consulte a [visão geral da arquitetura](docs/arquitetura/visao-geral.md).

## Requisitos

- Linux;
- Go 1.27.0 ou versão compatível com a definida em [`go.mod`](go.mod);
- MySQL 8.4;
- espaço em disco suficiente para os arquivos compactados e os schemas criados;
- `local_infile` habilitado no servidor MySQL;
- usuário MySQL com permissões compatíveis com a criação e a carga dos schemas.

Verifique a configuração atual do MySQL:

```sql
SHOW GLOBAL VARIABLES LIKE 'local_infile';
```

O valor precisa estar como `ON`.

Consulte [Preparação do MySQL](docs/operacao/mysql.md) para habilitar `local_infile`, criar o usuário operacional e, opcionalmente, iniciar uma instância MySQL com Docker Compose.

## Execução local

Clone o repositório e entre no diretório do projeto:

```bash
git clone https://github.com/danielsantello/go-cnpj-loader.git
cd go-cnpj-loader
```

Durante o desenvolvimento, execute diretamente pelo código-fonte:

```bash
go run ./cmd/cnpj-loader --help
```

Para uso recorrente, compile um binário local:

```bash
go build \
  -o bin/cnpj-loader \
  ./cmd/cnpj-loader
```

Consulte os comandos disponíveis:

```bash
bin/cnpj-loader --help
```

Nos exemplos seguintes, substitua `bin/cnpj-loader` por `go run ./cmd/cnpj-loader` quando quiser executar diretamente pelo código-fonte.

As instruções para gerar um binário com metadados de versão estão em [Compilação e execução local](docs/desenvolvimento/compilacao.md).

## Configuração

Crie um arquivo local a partir do modelo:

```bash
cp .env.example .env
```

Configuração disponível:

```dotenv
CNPJ_LOADER_MYSQL_HOST=127.0.0.1
CNPJ_LOADER_MYSQL_PORT=3306
CNPJ_LOADER_MYSQL_USER=
CNPJ_LOADER_MYSQL_PASSWORD=
CNPJ_LOADER_MYSQL_CONNECT_TIMEOUT=5s
CNPJ_LOADER_CONTROL_SCHEMA=cnpj_loader_control
```

O programa não lê o arquivo `.env` automaticamente. Para carregar suas variáveis somente durante uma execução, utilize um subshell:

```bash
(
  set -a
  source .env
  set +a

  bin/cnpj-loader version
)
```

O `.env` pode conter segredos, é ignorado pelo Git e não deve ser versionado.

## Download de uma publicação

O download não utiliza a conexão com o MySQL.

Informe o ano, o mês e um diretório absoluto de destino:

```bash
bin/cnpj-loader download \
  --reference-year 2026 \
  --reference-month 8 \
  --destination /dados/cnpj/2026-08
```

Arquivos existentes com o tamanho esperado são reaproveitados, e novos downloads utilizam arquivos temporários `.part`.

Para conhecer as regras de validação, reaproveitamento e tratamento de falhas, consulte o [guia de operação](docs/operacao/README.md).

## Carga de uma publicação

Carregue a configuração e informe o diretório que contém todos os ZIPs da publicação:

```bash
(
  set -a
  source .env
  set +a

  bin/cnpj-loader load \
    --reference-year 2026 \
    --reference-month 8 \
    --source /dados/cnpj/2026-08
)
```

O comando verifica a estrutura, a completude, o tamanho e o SHA-256 dos arquivos antes de acessar o banco. Em seguida, aplica as migrations de controle, cria uma nova versão e carrega cada arquivo em uma transação própria.

A versão recebe o status `ready` somente quando toda a carga termina sem erros nem warnings. Consulte o [guia de operação](docs/operacao/README.md) para conhecer o fluxo completo e o tratamento de falhas.

## Schema de controle

O comando abaixo cria, atualiza ou apenas valida o schema permanente de controle, sem iniciar uma carga:

```bash
(
  set -a
  source .env
  set +a

  bin/cnpj-loader migrate-control
)
```

Sua execução separada é opcional. O comando `load` aplica automaticamente as migrations pendentes antes de iniciar uma nova versão.

## Informações da versão

```bash
bin/cnpj-loader version
```

Exemplo de um binário local sem metadados injetados:

```text
Versão: dev
Commit: unknown
Compilado em: unknown
Go: go1.27.0
```

## Datasets carregados

Cada publicação completa contém os seguintes conjuntos de dados:

- atividades econômicas;
- motivos da situação cadastral;
- municípios;
- naturezas jurídicas;
- países;
- qualificações dos sócios;
- empresas;
- estabelecimentos;
- sócios;
- opções pelo Simples Nacional e MEI.

As tabelas e colunas utilizam nomes em inglês. As mensagens destinadas ao operador e a documentação são mantidas em português.

## Resultado comprovado

A carga integral da publicação de agosto de 2026 foi validada com os 37 arquivos oficiais:

| Métrica | Resultado |
| --- | ---: |
| Arquivos ZIP | 37 |
| Tamanho compactado | 7.692.070.350 bytes |
| Registros carregados | 220.339.195 |
| Tempo total | 58m13.785s |
| Warnings do MySQL | 0 |

O teste foi executado com streaming direto dos ZIPs e transações independentes por arquivo. O resultado pode variar conforme processador, armazenamento e configuração do MySQL.

## Estado atual

O núcleo de download e carga está implementado e foi validado com uma publicação completa.

A exclusão de schemas antigos e a limpeza dos arquivos baixados permanecem sob responsabilidade do operador.

## Documentação

- [Índice da documentação](docs/README.md)
- [Visão geral da arquitetura](docs/arquitetura/visao-geral.md)
- [Compilação e execução local](docs/desenvolvimento/compilacao.md)
- [Testes](docs/desenvolvimento/testes.md)
- [Decisões arquiteturais](docs/decisoes/README.md)
- [Estudos de carga](docs/estudos/carga/README.md)
- [Estudos de índices](docs/estudos/indices/README.md)
- [Estudos de busca textual](docs/estudos/busca-textual/README.md)

## Licença

Este projeto é distribuído sob a [licença MIT](LICENSE).
