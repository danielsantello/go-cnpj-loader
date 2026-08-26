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

## Teste de integração com MySQL

O teste de integração valida uma conexão real utilizando as configurações `CNPJ_LOADER_*`.

Por padrão, esse teste é ignorado. Isso permite executar a suíte comum e a integração contínua sem depender de uma instância externa do MySQL.

Antes da execução, crie o arquivo local `.env` a partir do modelo:

```bash
cp .env.example .env
```

Preencha no `.env` as credenciais e o endereço da instância MySQL utilizada no teste.

Execute o teste de integração com:

```bash
(set -a; source .env; set +a; CNPJ_LOADER_RUN_INTEGRATION_TESTS=1 go test -v ./internal/database -run '^TestMySQLConnection$')
```

O subshell delimitado por parênteses impede que as variáveis carregadas permaneçam exportadas no terminal depois da execução.

O arquivo `.env` contém configurações locais e possíveis segredos. Ele é ignorado pelo Git e nunca deve ser versionado.
