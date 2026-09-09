# Visão geral da arquitetura

## Responsabilidade

O `go-cnpj-loader` é uma aplicação de linha de comando responsável por baixar e carregar no MySQL os dados públicos de CNPJ disponibilizados pela Receita Federal.

O programa:

- consulta publicações remotas;
- baixa seus arquivos ZIP;
- valida publicações armazenadas em diretório;
- registra o histórico operacional;
- cria schemas de dados versionados;
- carrega todos os datasets de uma publicação.

O loader é executado sob demanda. Ele não funciona como serviço permanente.

## Limites

O loader não é responsável por:

- atender consultas de consumidores;
- conhecer ou alterar a configuração de uma API;
- publicar automaticamente uma versão;
- identificar qual schema uma aplicação externa utiliza;
- excluir automaticamente versões anteriores;
- corrigir ou validar semanticamente os dados da Receita;
- fornecer pesquisa textual por nomes;
- criar índices para consultas nesta etapa do projeto.

A publicação de uma versão para uma aplicação consumidora e a remoção de schemas anteriores são decisões externas ao loader.

## Princípio sobre os dados

A Receita Federal é considerada a fonte soberana dos dados cadastrais.

O loader não verifica se um CNPJ, CNAE, endereço ou outro conteúdo está semanticamente correto. Valores incomuns são preservados quando a estrutura técnica definida pelo projeto comporta esses valores.

A validação realizada é técnica. Ela inclui:

- estrutura e completude da publicação;
- classificação dos arquivos;
- tamanho dos arquivos;
- SHA-256 dos arquivos;
- quantidade e compatibilidade das colunas;
- erros retornados pelo MySQL;
- warnings produzidos durante a carga.

Qualquer warning do MySQL é tratado como erro para impedir truncamentos, conversões inválidas ou perdas silenciosas causadas pelo schema.

## Interface pública

A interface atual possui quatro comandos:

```text
cnpj-loader download
cnpj-loader load
cnpj-loader migrate-control
cnpj-loader version
```

`download` e `load` são independentes:

- `download` consulta e baixa uma publicação para um diretório absoluto;
- `load` utiliza exclusivamente os ZIPs existentes em um diretório;
- a carga nunca inicia um download implicitamente.

## Fluxo de download

O comando `download`:

1. valida o ano, o mês e o diretório de destino;
2. consulta a publicação remota por WebDAV;
3. interpreta a listagem de arquivos ZIP;
4. valida nomes, tamanhos e duplicidades;
5. ordena os arquivos;
6. cria o diretório de destino quando necessário;
7. processa um arquivo por vez;
8. reaproveita arquivos finais com o tamanho esperado;
9. baixa os demais para arquivos temporários terminados em `.part`;
10. confere a quantidade de bytes recebida;
11. renomeia o arquivo somente depois da conclusão.

O download é sequencial. Testes reais mostraram que o servidor remoto pode interromper transferências concorrentes.

O cliente HTTP limita a espera pelos cabeçalhos da resposta, mas não impõe um timeout global ao corpo. Arquivos grandes podem continuar sendo transferidos pelo tempo necessário.

## Fluxo de carga

O comando `load` executa duas fases principais.

### Verificação local

Antes de abrir uma conexão com o MySQL, o comando:

1. valida o ano e o mês de referência;
2. valida o diretório de origem;
3. carrega o catálogo incorporado de datasets;
4. descobre os ZIPs;
5. classifica e ordena os arquivos;
6. verifica a completude da publicação;
7. confere o tamanho e o SHA-256 de cada arquivo.

Uma publicação incompleta ou tecnicamente inválida não inicia uma carga.

### Persistência e carga

Depois da verificação local, o comando:

1. abre a conexão com o MySQL;
2. verifica a disponibilidade da conexão;
3. cria ou atualiza o schema de controle;
4. registra a publicação e seus arquivos;
5. cria uma versão com estado `pending`;
6. cria o novo schema de dados;
7. cria as tabelas sem chaves e índices;
8. altera a versão para `loading`;
9. transmite o conteúdo de cada ZIP diretamente para o MySQL;
10. carrega cada arquivo em uma transação independente;
11. verifica os warnings antes de cada `COMMIT`;
12. marca a versão como `ready` depois da conclusão integral.

Os arquivos CSV não são extraídos fisicamente. O loader abre o único conteúdo de cada ZIP e o fornece ao driver MySQL por um leitor registrado especificamente para aquela carga.

## Estratégia de carga

A importação em massa utiliza `LOAD DATA LOCAL INFILE`.

O driver recebe o conteúdo por `RegisterReaderHandler`. O programa não habilita acesso irrestrito a arquivos locais com `allowAllFiles=true`.

A configuração do CSV utiliza:

```text
FIELDS TERMINATED BY ';'
ENCLOSED BY '"'
ESCAPED BY ''
```

`ESCAPED BY ''` é necessário porque existem valores da Receita terminados em barra invertida.

Os dados de entrada são interpretados como `latin1` e armazenados em tabelas `utf8mb4` com collation `utf8mb4_0900_ai_ci`.

Transformações técnicas incluem:

- datas sentinelas vazias, `0` e `00000000` convertidas em `NULL`;
- campos opcionais vazios convertidos em `NULL` conforme o layout;
- capital social convertido de vírgula decimal para `DECIMAL(20,2)`;
- CNPJ completo derivado de raiz, ordem do estabelecimento e dígitos verificadores.

O loader não modifica semanticamente o conteúdo cadastral.

## Transações e warnings

A carga completa não constitui uma única transação.

Cada arquivo é uma unidade transacional independente:

1. o loader inicia a transação;
2. registra um leitor para o conteúdo do ZIP;
3. executa `LOAD DATA LOCAL INFILE`;
4. consulta `SHOW COUNT(*) WARNINGS`;
5. executa `COMMIT` somente quando não existem warnings.

Quando há warnings, o loader consulta até dez detalhes com `SHOW WARNINGS`, executa o rollback do arquivo e retorna um erro.

Essa estratégia limita o escopo de rollback sem prometer uma transação global que o DDL do MySQL não poderia garantir.

## Ciclo de vida das versões

Uma versão bem-sucedida percorre:

```text
pending -> loading -> ready
```

Uma versão que falha durante a carga percorre:

```text
pending -> loading -> failed
```

Ao marcar uma versão como `ready`, o loader registra `ready_at_utc` e `status_changed_at_utc` com o mesmo instante.

Ao marcar uma versão como `failed`, registra `failed_at_utc` e `status_changed_at_utc`.

Depois que a versão entra em `loading`, uma finalização diferida tenta registrar a falha mesmo quando o contexto principal foi cancelado. Essa tentativa utiliza um contexto não cancelado com uma janela limitada e preserva o erro original.

Quando ocorre uma falha impeditiva:

- o processamento é interrompido;
- a versão é marcada como `failed` quando possível;
- o schema incompleto é preservado;
- os arquivos são preservados;
- versões anteriores não são alteradas;
- nenhuma limpeza automática é realizada.

## Schemas de dados

Cada carga cria um schema novo. Um schema existente nunca é truncado, reutilizado ou sobrescrito silenciosamente.

A convenção atual é:

```text
cnpj_YYYY_MM_NNN
```

Exemplo:

```text
cnpj_2026_08_001
```

A sequência diferencia novas cargas da mesma publicação dentro de um ambiente lógico.

O cálculo atual da sequência considera a publicação e o ambiente, mas o nome físico do schema ainda não inclui o ambiente. Isso pode produzir colisões quando a mesma publicação é carregada em ambientes diferentes usando a mesma instância MySQL.

A nomenclatura será revisada antes da primeira release.

## Tabelas de dados

Cada schema versionado possui:

```text
schema_metadata
economic_activities
registration_status_reasons
municipalities
legal_natures
countries
partner_qualifications
companies
establishments
partners
simple_tax_options
```

`schema_metadata` registra:

- versão do formato do schema;
- ano de referência;
- mês de referência;
- data de criação em UTC.

Uma aplicação consumidora pode verificar a compatibilidade do schema sem depender do schema de controle.

As tabelas são criadas sem chaves e índices durante a carga. A estratégia de indexação será definida posteriormente por consultas reais e benchmarks reproduzíveis.

## Schema de controle

O schema permanente configurado por `CNPJ_LOADER_CONTROL_SCHEMA` mantém o histórico operacional.

O valor padrão é:

```text
cnpj_loader_control
```

As migrations atuais criam estruturas para:

- histórico das migrations;
- publicações;
- arquivos das publicações;
- versões;
- execuções;
- configurações das execuções;
- etapas das execuções;
- cargas de arquivos;
- eventos das execuções.

Ao contrário das tabelas de dados durante a carga, o schema de controle possui as restrições e os índices necessários à sua própria consistência.

O comando `migrate-control` cria, atualiza ou valida esse schema sem iniciar uma carga. O comando `load` também executa automaticamente as migrations pendentes.

## Componentes

### `internal/buildinfo`

Mantém versão, commit, data de compilação e versão do runtime Go.

### `internal/cli`

Define os comandos públicos e coordena seus fluxos.

### `internal/config`

Lê e valida as variáveis de ambiente.

### `internal/database`

Configura a conexão MySQL e verifica sua disponibilidade.

### `internal/download`

Consulta o compartilhamento WebDAV da Receita e realiza o download sequencial dos ZIPs.

### `internal/publication`

Mantém o catálogo incorporado, descobre arquivos, classifica datasets, valida a completude, calcula fingerprints e abre o conteúdo dos ZIPs.

### `internal/control`

Aplica migrations e registra publicações, arquivos, versões e seus estados.

### `internal/data`

Cria o schema de dados, cria as tabelas e executa as cargas dos dez datasets.

## Concorrência

O download atual é sequencial.

A carga atual também processa os arquivos de maneira sequencial, utilizando uma transação por arquivo.

O bloqueio consultivo para impedir operações mutáveis concorrentes ainda não está implementado. Até que essa proteção exista, o operador não deve iniciar cargas ou migrations concorrentes contra o mesmo schema de controle.

## Independência dos consumidores

O loader não conhece nenhuma API consumidora.

Uma API pode utilizar qualquer schema compatível, independentemente de ele ter sido produzido por este loader. Da mesma forma, uma versão `ready` está tecnicamente concluída, mas não é publicada automaticamente para consumidor algum.

A troca do schema utilizado por uma aplicação é uma operação externa e explícita.

## Idioma

O projeto utiliza:

- português na documentação;
- português nas mensagens destinadas ao operador;
- inglês no código Go;
- inglês nos comandos e flags;
- inglês nos schemas, tabelas e colunas;
- inglês nos códigos e valores internos;
- conteúdo da Receita preservado conforme a fonte.

Elementos convencionais gerados pelo Cobra, como `Usage`, `Flags` e `Help`, permanecem em inglês.

## Pendências arquiteturais

Antes da primeira release:

- revisar a nomenclatura dos schemas entre ambientes;
- definir e documentar os privilégios mínimos do usuário MySQL;
- documentar a configuração persistente de `local_infile`;
- adicionar testes automatizados ao pacote de download.

Melhorias posteriores incluem:

- bloqueio consultivo para operações mutáveis;
- comandos de consulta e exclusão de versões;
- limpeza explícita de arquivos;
- índices configuráveis;
- retomada de downloads com HTTP Range;
- progresso de download por bytes;
- reaproveitamento de conteúdo já validado;
- busca automática da publicação mais recente;
- otimizações do redo log;
- mecanismo de busca textual.
