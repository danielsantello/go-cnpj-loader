# go-cnpj-loader

Carregador versionado dos dados públicos de CNPJ disponibilizados pela Receita Federal, desenvolvido em Go para processar grandes volumes de dados no MySQL.

O programa baixa e carrega publicações completas, registra o histórico operacional e cria um novo schema de dados para cada versão. Ele é independente de qualquer API ou aplicação consumidora.

> O projeto está em fase de preparação da primeira release pública.

## Funcionalidades

- consulta dos arquivos de uma publicação da Receita Federal;
- download sequencial dos arquivos ZIP;
- reaproveitamento de arquivos existentes quando o tamanho corresponde ao arquivo remoto;
- uso de arquivos temporários `.part` durante downloads;
- validação da estrutura e da completude da publicação;
- verificação de tamanho e SHA-256 dos arquivos;
- criação de schemas de dados versionados;
- carga dos dez datasets da publicação;
- transmissão direta do conteúdo dos ZIPs para o MySQL, sem extração física dos CSVs;
- importação em massa com `LOAD DATA LOCAL INFILE`;
- transação independente para cada arquivo;
- tratamento de warnings do MySQL como erro;
- registro de publicações, arquivos, versões e migrations em um schema permanente de controle;
- preservação do schema incompleto e do histórico em caso de falha;
- identificação do binário por versão, commit, data de compilação e versão do Go.

## Princípios

- a Receita Federal é considerada a fonte soberana dos dados;
- não são realizadas correções ou validações semânticas do conteúdo cadastral;
- cada carga cria um novo schema;
- schemas existentes nunca são sobrescritos ou truncados silenciosamente;
- versões anteriores nunca são removidas automaticamente;
- downloads e cargas são operações independentes;
- os CSVs não são extraídos no disco;
- cada arquivo constitui uma unidade transacional;
- qualquer warning do MySQL interrompe a operação correspondente;
- as tabelas são carregadas sem chaves e índices;
- índices serão definidos posteriormente, com base em consultas e experimentos reais;
- a publicação de uma versão para uma aplicação consumidora é uma decisão externa ao loader.

## Como funciona

O fluxo normal possui duas operações independentes:

1. `download` consulta e baixa os arquivos ZIP de uma publicação para um diretório absoluto;
2. `load` descobre e valida os ZIPs existentes em um diretório, cria um novo schema versionado e carrega os dados no MySQL.

O comando `load` nunca inicia um download implicitamente.

Durante a carga, o conteúdo de cada ZIP é transmitido diretamente ao MySQL. Cada arquivo é carregado em sua própria transação e somente recebe `COMMIT` quando termina sem erros nem warnings.

Uma carga bem-sucedida produz um schema como:

```text
cnpj_2026_08_001
```

O schema permanente de controle registra a publicação, os arquivos e o ciclo de vida da versão:

```text
pending -> loading -> ready
```

Quando ocorre uma falha:

```text
pending -> loading -> failed
```

O schema incompleto e os arquivos são preservados para diagnóstico. Nenhuma limpeza automática é realizada.

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

A definição dos privilégios mínimos e a configuração persistente de `local_infile` serão concluídas antes da primeira release.

## Compilação

Clone o repositório e entre no diretório do projeto:

```bash
git clone https://github.com/danielsantello/go-cnpj-loader.git
cd go-cnpj-loader
```

Compile o binário:

```bash
go build \
  -o bin/cnpj-loader \
  ./cmd/cnpj-loader
```

Consulte os comandos disponíveis:

```bash
bin/cnpj-loader --help
```

As instruções para gerar um binário com metadados de versão estão em [Compilação e execução local](docs/desenvolvimento/compilacao.md).

## Configuração

Crie um arquivo local a partir do modelo:

```bash
cp .env.example .env
```

Configuração disponível:

```dotenv
CNPJ_LOADER_ENVIRONMENT=development
CNPJ_LOADER_MYSQL_HOST=127.0.0.1
CNPJ_LOADER_MYSQL_PORT=3306
CNPJ_LOADER_MYSQL_USER=
CNPJ_LOADER_MYSQL_PASSWORD=
CNPJ_LOADER_MYSQL_CONNECT_TIMEOUT=5s
CNPJ_LOADER_CONTROL_SCHEMA=cnpj_loader_control
```

Ambientes lógicos aceitos:

- `development`;
- `benchmark`;
- `production`.

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

Os arquivos são processados sequencialmente. Essa escolha evita sobrecarregar o servidor remoto durante o download de publicações grandes.

Quando um arquivo final já existe e possui o tamanho esperado, ele é reaproveitado. Se o tamanho for diferente, o programa falha sem sobrescrevê-lo.

Um novo download é gravado inicialmente como `<nome>.part` e somente é renomeado para o nome definitivo depois de ser concluído e conferido.

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

Antes de acessar o banco, o comando:

1. valida o ano, o mês e o diretório;
2. descobre e classifica os arquivos;
3. verifica a completude da publicação;
4. confere o tamanho e o SHA-256 de cada arquivo.

Em seguida, ele:

1. conecta-se ao MySQL;
2. aplica as migrations pendentes do schema de controle;
3. registra a publicação e seus arquivos;
4. cria uma nova versão e seu schema de dados;
5. cria as tabelas sem índices;
6. carrega cada arquivo em uma transação própria;
7. marca a versão como `ready` quando toda a carga termina sem warnings.

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
| Tempo total | 53m8.493s |
| Warnings do MySQL | 0 |

O teste foi executado com streaming direto dos ZIPs e transações independentes por arquivo. O resultado pode variar conforme processador, armazenamento e configuração do MySQL.

## Estado atual

O núcleo de download e carga está implementado e foi validado com uma publicação completa.

Antes da primeira release ainda serão concluídos:

- documentação operacional;
- configuração persistente de `local_infile`;
- definição dos privilégios mínimos do usuário MySQL;
- revisão da nomenclatura dos schemas;
- testes automatizados do pacote de download;
- validação final e geração do binário rastreável.

Comandos de exclusão, limpeza e consulta de versões ainda não fazem parte da interface pública.

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
