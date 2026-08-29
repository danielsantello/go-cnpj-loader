# Testes

O projeto possui testes unitários e testes de integração.

## Testes unitários

A suíte completa pode ser executada com:

```bash
go test ./...
```

Para utilizar o detector de concorrência:

```bash
go test -race ./...
```

Os testes unitários não exigem uma instância do MySQL.

Os testes de integração são ignorados por padrão. Isso permite executar a suíte comum e a integração contínua sem depender de uma instância externa.

## Configuração dos testes de integração

Crie o arquivo local `.env` a partir do modelo:

```bash
cp .env.example .env
```

Preencha no `.env` as credenciais e o endereço da instância MySQL utilizada nos testes.

O arquivo `.env` contém configurações locais e possíveis segredos. Ele é ignorado pelo Git e nunca deve ser versionado.

Os comandos abaixo utilizam um subshell delimitado por parênteses. As variáveis carregadas deixam de existir no terminal quando a execução termina.

## Teste de conexão com o MySQL

Este teste:

- carrega e valida a configuração;
- abre o pool de conexões;
- executa um ping real;
- fecha o pool ao terminar.

Execute:

```bash
(
  set -a
  source .env
  set +a

  CNPJ_LOADER_RUN_INTEGRATION_TESTS=1 \
    go test -v \
      ./internal/database \
      -run '^TestMySQLConnection$'
)
```

A credencial precisa apenas conseguir conectar-se à instância para realizar esse teste.

## Teste das migrations do schema de controle

Este teste percorre o fluxo completo:

1. carrega e valida a configuração;
2. conecta-se ao MySQL;
3. carrega o catálogo de migrations incorporado ao executável de teste;
4. cria um schema temporário;
5. executa todas as migrations disponíveis;
6. executa novamente o mecanismo para validar sua idempotência;
7. verifica as tabelas `control_schema_migrations`, `publications`, `publication_files`, `versions`, `executions` e `execution_configurations`;
8. compara todos os registros do histórico com o catálogo incorporado;
9. provoca uma falha controlada de DDL;
10. verifica o status `failed`, o código MySQL, o SQL state, a mensagem e o horário da falha;
11. remove o schema temporário com `t.Cleanup`.

Execute:

```bash
(
  set -a
  source .env
  set +a

  CNPJ_LOADER_RUN_INTEGRATION_TESTS=1 \
    go test -v \
      ./internal/control/migrations \
      -run '^TestBootstrapCreatesControlSchema$'
)
```

A falha controlada faz parte do teste e é capturada pelo próprio código. Quando todas as verificações passam, o resultado final do teste continua sendo `PASS`.

A credencial usada nesse teste precisa possuir privilégios para criar, alterar e excluir objetos no schema temporário:

```text
cnpj_loader_migrations_integration_test
```

Utilize somente uma instância isolada de desenvolvimento. Não use credenciais operacionais de produção para executar testes de integração.

Antes de começar, o teste confirma que o schema temporário não existe. Se encontrar um schema com esse nome, ele interrompe sem excluí-lo automaticamente, permitindo investigação.

Depois que o teste começa a criar sua estrutura, registra uma função com `t.Cleanup`. Essa função tenta remover exclusivamente o schema temporário mesmo quando uma verificação termina com `t.Fatal`.

Uma interrupção abrupta do processo ou da máquina pode impedir o cleanup. Nesse caso, inspecione o schema preservado antes de removê-lo manualmente.

## Execução dos dois testes de integração

Quando a credencial possuir todos os privilégios necessários aos dois cenários:

```bash
(
  set -a
  source .env
  set +a

  CNPJ_LOADER_RUN_INTEGRATION_TESTS=1 \
    go test -v \
      ./internal/database \
      ./internal/control/migrations
)
```
