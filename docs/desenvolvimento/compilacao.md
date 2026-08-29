# Compilação e execução local

Este guia descreve como compilar e executar o `cnpj-loader` diretamente em um ambiente Linux.

O uso de releases prontas e a execução com Docker serão documentados separadamente quando esses fluxos estiverem implementados.

## Pré-requisitos

- Go na versão definida em `go.mod`;
- acesso a uma instância MySQL compatível;
- credenciais com os privilégios necessários para a operação executada.

Consulte a versão instalada do Go:

```bash
go version
```

## Compilação simples

Para gerar um binário local de desenvolvimento:

```bash
go build \
  -o bin/cnpj-loader \
  ./cmd/cnpj-loader
```

Sem metadata injetado, o comando `version` apresentará os valores padrão:

```text
Versão: dev
Commit: unknown
Compilado em: unknown
```

A versão do runtime Go é obtida automaticamente durante a execução.

## Compilação com metadata

Para produzir um binário local rastreável:

```bash
build_version="v0.1.0-dev"
build_commit="$(git rev-parse --short HEAD)"
build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

Compile utilizando os valores definidos:

```bash
go build \
  -trimpath \
  -ldflags "\
-X github.com/danielsantello/go-cnpj-loader/internal/buildinfo.Version=${build_version} \
-X github.com/danielsantello/go-cnpj-loader/internal/buildinfo.Commit=${build_commit} \
-X github.com/danielsantello/go-cnpj-loader/internal/buildinfo.BuildDate=${build_date}" \
  -o bin/cnpj-loader \
  ./cmd/cnpj-loader
```

A opção `-trimpath` remove dos metadados de compilação os caminhos absolutos do ambiente local.

A opção `-ldflags` injeta no executável:

- versão do programa;
- commit utilizado;
- data da compilação em UTC.

Os arquivos SQL declarados com `go:embed` são incorporados ao executável. A pasta de migrations não precisa acompanhar o binário.

## Verificação do binário

Consulte o formato do arquivo:

```bash
file bin/cnpj-loader
```

Exiba os comandos disponíveis:

```bash
bin/cnpj-loader --help
```

Confira o metadata:

```bash
bin/cnpj-loader version
```

## Configuração local

Crie o arquivo privado de configuração a partir do modelo:

```bash
cp .env.example .env
```

Preencha as variáveis `CNPJ_LOADER_*` conforme a instância MySQL utilizada.

O programa não lê `.env` automaticamente. O arquivo precisa ser carregado no ambiente do processo.

Para executar em um subshell temporário:

```bash
(
  set -a
  source .env
  set +a

  bin/cnpj-loader migrate-control
)
```

Os parênteses criam um subshell. As variáveis exportadas deixam de existir quando a execução termina.

O arquivo `.env` contém configurações locais e possíveis segredos. Ele é ignorado pelo Git e nunca deve ser versionado.

## Criação e atualização do schema de controle

O comando:

```bash
bin/cnpj-loader migrate-control
```

carrega o catálogo de migrations incorporado ao executável e verifica o estado do schema configurado.

Quando o schema ainda não existe, o comando:

1. cria o schema de controle;
2. executa todas as migrations incorporadas em ordem;
3. registra cada aplicação no histórico.

Quando o schema já existe, o comando:

1. consulta `control_schema_migrations`;
2. valida versão, nome, checksum e status das migrations registradas;
3. executa somente as migrations pendentes;
4. preserva a metadata das migrations aplicadas anteriormente.

Quando não existem migrations pendentes, o comando apenas valida o histórico e termina com sucesso.

Saída esperada:

```text
Schema de controle "cnpj_loader_control" atualizado com sucesso.
```

A execução é interrompida quando o histórico está inconsistente, contém uma migration desconhecida, possui checksum ou nome divergente, ou apresenta status diferente de `applied`.

A primeira migration cria a própria tabela de histórico e, por isso, só pode ser registrada depois de sua execução. A partir da segunda migration, cada aplicação é registrada inicialmente como `applying` e termina como `applied` ou `failed`.

## Teste manual isolado

Para testar o binário sem utilizar o nome definitivo do schema de controle, sobrescreva a variável somente no subshell:

```bash
(
  set -a
  source .env
  set +a

  export CNPJ_LOADER_CONTROL_SCHEMA=cnpj_loader_manual_test

  bin/cnpj-loader migrate-control
)
```

Depois da execução, inspecione no MySQL:

```sql
SELECT
    version,
    name,
    status,
    loader_version,
    loader_commit,
    started_at_utc,
    finished_at_utc
FROM cnpj_loader_manual_test.control_schema_migrations
ORDER BY version;
```

Quando a inspeção terminar, exclua exclusivamente o schema criado para esse teste:

```sql
DROP DATABASE cnpj_loader_manual_test;
```

Nunca utilize nesse procedimento o nome de um schema que contenha dados necessários.

## Privilégios

Não utilize `root` como credencial operacional do loader.

O conjunto mínimo de privilégios MySQL será definido e documentado antes da primeira versão operacional. Enquanto essa definição não estiver concluída, utilize somente uma instância isolada de desenvolvimento e credenciais próprias para testes.
