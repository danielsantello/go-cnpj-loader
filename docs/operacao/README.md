# Operação

Este guia descreve a preparação e a execução dos comandos públicos do `cnpj-loader`.

O programa possui quatro comandos:

```text
cnpj-loader download
cnpj-loader load
cnpj-loader migrate-control
cnpj-loader version
```

`download` e `load` são operações independentes. A carga nunca inicia um download implicitamente.

## Preparação do binário

Compile o projeto:

```bash
go build \
  -o bin/cnpj-loader \
  ./cmd/cnpj-loader
```

Confirme os comandos disponíveis:

```bash
bin/cnpj-loader --help
```

Consulte [Compilação e execução local](../desenvolvimento/compilacao.md) para gerar um binário com metadados de versão, commit e data de compilação.

## Configuração do MySQL

A carga requer uma instância MySQL 8.4 com `local_infile` habilitado.

Verifique o valor atual:

```sql
SHOW GLOBAL VARIABLES LIKE 'local_infile';
```

O resultado esperado é:

```text
local_infile | ON
```

O loader utiliza `LOAD DATA LOCAL INFILE` para transmitir ao MySQL o conteúdo dos arquivos mantidos dentro dos ZIPs. Os CSVs não são extraídos fisicamente.

A configuração persistente de `local_infile` e o conjunto mínimo de privilégios do usuário operacional serão documentados antes da primeira release.

## Configuração do loader

Crie o arquivo local de configuração:

```bash
cp .env.example .env
```

Preencha as variáveis:

```dotenv
CNPJ_LOADER_ENVIRONMENT=development
CNPJ_LOADER_MYSQL_HOST=127.0.0.1
CNPJ_LOADER_MYSQL_PORT=3306
CNPJ_LOADER_MYSQL_USER=
CNPJ_LOADER_MYSQL_PASSWORD=
CNPJ_LOADER_MYSQL_CONNECT_TIMEOUT=5s
CNPJ_LOADER_CONTROL_SCHEMA=cnpj_loader_control
```

`CNPJ_LOADER_MYSQL_USER` e `CNPJ_LOADER_MYSQL_PASSWORD` não possuem valores padrão.

Os ambientes lógicos aceitos são:

- `development`;
- `benchmark`;
- `production`.

O programa não lê `.env` automaticamente. Carregue o arquivo no ambiente do processo:

```bash
(
  set -a
  source .env
  set +a

  bin/cnpj-loader version
)
```

Os parênteses criam um subshell. As variáveis exportadas deixam de existir quando a execução termina.

O arquivo `.env` pode conter segredos, é ignorado pelo Git e nunca deve ser versionado.

## Consultar uma versão do binário

Execute:

```bash
bin/cnpj-loader version
```

A saída contém:

- versão do loader;
- commit utilizado na compilação;
- data da compilação em UTC;
- versão do runtime Go.

Um binário local sem metadados injetados apresenta valores de desenvolvimento:

```text
Versão: dev
Commit: unknown
Compilado em: unknown
Go: go1.27.0
```

## Baixar uma publicação

O comando `download` consulta os arquivos de uma publicação da Receita Federal e os grava em um diretório absoluto.

Ele não utiliza a configuração nem a conexão com o MySQL.

Exemplo:

```bash
bin/cnpj-loader download \
  --reference-year 2026 \
  --reference-month 8 \
  --destination /dados/cnpj/2026-08
```

As três flags são obrigatórias:

| Flag | Descrição |
| --- | --- |
| `--reference-year` | Ano de referência da publicação |
| `--reference-month` | Mês de referência da publicação |
| `--destination` | Diretório absoluto que receberá os ZIPs |

### Comportamento do download

O comando:

1. consulta a listagem remota da publicação;
2. identifica e ordena os arquivos ZIP;
3. cria o diretório de destino quando necessário;
4. processa um arquivo por vez;
5. reaproveita arquivos finais com o tamanho esperado;
6. baixa os demais arquivos para nomes temporários terminados em `.part`;
7. confere a quantidade de bytes recebida;
8. renomeia cada arquivo somente depois da conclusão.

O download é sequencial porque o servidor remoto apresentou instabilidade durante testes concorrentes.

Se um arquivo final já existir com tamanho diferente do esperado, o comando falhará sem sobrescrevê-lo.

Em falhas tratadas, o arquivo `.part` correspondente é removido. Um arquivo incompleto nunca é apresentado como ZIP concluído.

Exemplo de resumo:

```text
Publicação disponível em "/dados/cnpj/2026-08": 37 arquivos, 37 baixados, 0 reaproveitados, 7692070350 bytes.
```

Uma nova execução sobre o diretório completo reaproveita os arquivos existentes:

```text
Publicação disponível em "/dados/cnpj/2026-08": 37 arquivos, 0 baixados, 37 reaproveitados, 7692070350 bytes.
```

## Preparar o schema de controle

O schema de controle mantém o histórico operacional do loader.

Para criá-lo ou atualizá-lo sem iniciar uma carga:

```bash
(
  set -a
  source .env
  set +a

  bin/cnpj-loader migrate-control
)
```

O comando:

1. conecta-se ao MySQL;
2. cria o schema de controle quando ele ainda não existe;
3. valida o histórico das migrations aplicadas;
4. executa somente as migrations pendentes;
5. termina sem alterações quando o schema já está atualizado.

Saída esperada:

```text
Schema de controle "cnpj_loader_control" atualizado com sucesso.
```

A execução explícita de `migrate-control` é opcional antes de uma carga. O comando `load` executa a mesma verificação e aplica migrations pendentes automaticamente.

## Carregar uma publicação

O comando `load` utiliza exclusivamente os ZIPs presentes no diretório informado. Ele não acessa o servidor da Receita Federal para baixar arquivos.

Exemplo:

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

As três flags são obrigatórias:

| Flag | Descrição |
| --- | --- |
| `--reference-year` | Ano de referência da publicação |
| `--reference-month` | Mês de referência da publicação |
| `--source` | Diretório absoluto que contém os ZIPs |

### Verificação da publicação

Antes de acessar o MySQL, o comando:

1. valida o ano e o mês;
2. valida o diretório de origem;
3. descobre e classifica os ZIPs;
4. verifica se todos os datasets e fragmentos esperados estão presentes;
5. confere o tamanho e o SHA-256 de cada arquivo.

Uma publicação incompleta ou tecnicamente inválida não inicia uma carga.

### Execução da carga

Depois da verificação, o comando:

1. conecta-se ao MySQL;
2. cria ou atualiza o schema de controle;
3. registra a publicação e seus arquivos;
4. cria uma nova versão;
5. cria um novo schema de dados;
6. cria as tabelas sem chaves e índices;
7. transmite o conteúdo dos ZIPs diretamente para o MySQL;
8. carrega cada arquivo em uma transação independente;
9. verifica os warnings antes de cada `COMMIT`;
10. marca a versão como `ready` após a conclusão integral.

Os dez datasets carregados são:

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

## Ciclo de vida das versões

Uma carga bem-sucedida percorre:

```text
pending -> loading -> ready
```

Uma carga interrompida por erro percorre:

```text
pending -> loading -> failed
```

Quando ocorre uma falha impeditiva:

- nenhuma nova carga de arquivo é iniciada;
- a versão é marcada como `failed`;
- o erro original é preservado;
- o schema incompleto é mantido;
- os arquivos da publicação são mantidos;
- nenhuma versão anterior é alterada;
- nenhuma limpeza automática é realizada.

O loader não sobrescreve, trunca ou exclui automaticamente um schema existente.

## Warnings do MySQL

Qualquer warning produzido durante `LOAD DATA LOCAL INFILE` é tratado como erro.

Antes do `COMMIT`, o loader consulta a contagem de warnings. Quando encontra algum:

1. consulta os detalhes disponíveis;
2. executa o rollback da transação do arquivo;
3. interrompe a carga;
4. marca a versão como `failed`.

Essa regra impede que truncamentos, conversões inválidas ou perdas causadas pelo schema passem silenciosamente.

## Resultado de referência

A publicação de agosto de 2026 foi processada integralmente:

| Métrica | Resultado |
| --- | ---: |
| Arquivos ZIP | 37 |
| Tamanho compactado | 7.692.070.350 bytes |
| Registros carregados | 220.339.195 |
| Tempo total | 53m8.493s |
| Warnings do MySQL | 0 |

O tempo varia conforme processador, armazenamento e configuração do MySQL.

## Limitações atuais

Ainda não existem comandos públicos para:

- consultar publicações ou versões registradas;
- excluir schemas de versões;
- limpar arquivos baixados;
- publicar uma versão para uma API;
- retomar uma carga interrompida.

As tabelas de dados ainda não recebem índices. A estratégia será definida posteriormente a partir de consultas reais e benchmarks reproduzíveis.
