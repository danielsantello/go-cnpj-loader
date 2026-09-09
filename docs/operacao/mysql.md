# Preparação do MySQL

Este documento descreve os requisitos do MySQL para executar o `cnpj-loader`.

O projeto é independente da forma de instalação do banco. O MySQL pode ser executado nativamente, em container ou em um serviço gerenciado, desde que satisfaça o contrato descrito aqui.

## Requisitos

O ambiente validado utiliza:

```text
MySQL 8.4
```

O servidor precisa:

- aceitar conexões do processo do loader;
- utilizar credenciais próprias para o loader;
- permitir a criação de schemas e tabelas;
- manter `local_infile` habilitado;
- possuir espaço suficiente para os schemas versionados e os arquivos temporários utilizados pelo MySQL.

O loader não deve utilizar a conta `root` como credencial operacional.

## `LOAD DATA LOCAL INFILE`

O loader utiliza:

```sql
LOAD DATA LOCAL INFILE
```

Com `LOCAL`, o conteúdo é lido pelo cliente e transmitido pela conexão ao servidor. No `go-cnpj-loader`, o cliente é o próprio processo Go.

Os CSVs não precisam existir no sistema de arquivos do servidor MySQL. O loader abre o conteúdo de cada ZIP e registra um leitor específico no driver.

O programa não utiliza:

```text
allowAllFiles=true
```

Somente leitores registrados explicitamente podem fornecer conteúdo ao driver.

## Verificar `local_infile`

Consulte o valor atual:

```sql
SHOW GLOBAL VARIABLES LIKE 'local_infile';
```

O resultado precisa indicar:

```text
local_infile | ON
```

Quando a variável está desabilitada, o servidor rejeita `LOAD DATA LOCAL INFILE`.

O suporte também precisa estar habilitado no cliente. O driver utilizado pelo projeto é configurado pelo próprio loader para aceitar os leitores registrados.

## Habilitação persistente

A forma de habilitar `local_infile` depende de como o MySQL é administrado.

### Arquivo de configuração

Em uma instalação que utiliza `my.cnf`, adicione a opção à seção do servidor:

```ini
[mysqld]
local_infile=ON
```

Reinicie o servidor e verifique novamente o valor efetivo.

A localização do arquivo e o procedimento de reinício variam conforme o sistema operacional e a instalação.

### Linha de comando do servidor

Quando o processo é iniciado diretamente, utilize:

```bash
mysqld --local-infile=ON
```

### Docker Compose

Em um serviço baseado na imagem oficial do MySQL, a opção pode ser declarada no `command`:

```yaml
services:
  mysql:
    image: mysql:8.4
    command: ["mysqld", "--local-infile=ON"]
```

Quando a configuração de um container existente muda, o Compose precisa recriá-lo para aplicar a nova linha de comando. Um volume persistente deve ser utilizado para preservar os bancos.

## Exemplo opcional com Docker Compose

O exemplo abaixo fornece um MySQL local para desenvolvimento. Docker não é requisito do loader.

Crie um `compose.yaml` fora do repositório ou em uma área privada de infraestrutura:

```yaml
name: cnpj-loader-local

services:
  mysql:
    image: mysql:8.4
    command: ["mysqld", "--local-infile=ON"]
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
    ports:
      - "127.0.0.1:3307:3306"
    volumes:
      - mysql-data:/var/lib/mysql

volumes:
  mysql-data:
```

Crie um arquivo `.env` no mesmo diretório:

```dotenv
MYSQL_ROOT_PASSWORD=substitua-por-uma-senha-segura
```

Esse arquivo contém segredo e não deve ser versionado.

Inicie o serviço:

```bash
docker compose up -d mysql
```

Confirme o estado:

```bash
docker compose ps
```

Verifique `local_infile` usando uma sessão administrativa:

```sql
SHOW GLOBAL VARIABLES LIKE 'local_infile';
```

A publicação da porta em `127.0.0.1` limita o acesso ao próprio host. Ajustes de rede para outros ambientes são responsabilidade do operador.

A conta `root` serve apenas para a administração inicial. Crie uma conta separada para o loader antes de executar migrations ou cargas.

## Privilégios mínimos atuais

As funcionalidades implementadas foram validadas no MySQL 8.4 com:

```text
CREATE
SELECT
INSERT
UPDATE
REFERENCES
```

| Privilégio | Uso pelo loader |
| --- | --- |
| `CREATE` | Criar schemas e tabelas |
| `SELECT` | Consultar migrations, publicações e versões |
| `INSERT` | Registrar o controle, inserir metadados e carregar os datasets |
| `UPDATE` | Atualizar migrations e estados das versões |
| `REFERENCES` | Criar as chaves estrangeiras do schema de controle |

O conjunto foi validado com as nove migrations atuais e com a criação, inserção e consulta em um schema de dados isolado.

## Privilégios não necessários

As funcionalidades atuais não exigem:

| Privilégio | Motivo |
| --- | --- |
| `FILE` | `LOAD DATA LOCAL` lê o conteúdo pelo cliente |
| `DROP` | O loader não exclui schemas nem tabelas |
| `DELETE` | O loader não exclui registros |
| `ALTER` | As migrations atuais criam estruturas novas |
| `INDEX` | Índices declarados em `CREATE TABLE` são cobertos por `CREATE` |
| `CREATE USER` | Contas são administradas externamente |
| `GRANT OPTION` | O loader não concede privilégios |

Novos comandos ou migrations podem alterar esse conjunto. Os privilégios devem ser reavaliados quando o programa passar a executar novas categorias de instruções SQL.

## Conta operacional

O host permitido para a conta depende da rede utilizada.

O exemplo abaixo utiliza `%` para admitir conexões provenientes de qualquer endereço que alcance o MySQL:

```sql
CREATE USER 'cnpj_loader'@'%'
    IDENTIFIED BY 'substitua-por-uma-senha-segura';
```

Em ambientes controlados, substitua `%` por um host ou padrão de rede mais restrito.

Não reutilize a senha administrativa do MySQL.

## Concessão para os schemas do loader

Os schemas atuais utilizam nomes iniciados por:

```text
cnpj_
```

No MySQL 8.4, privilégios em nível de banco podem utilizar `%` e `_` como curingas. Para tratar o sublinhado de `cnpj_` como caractere literal, ele precisa ser escapado:

```sql
GRANT CREATE, SELECT, INSERT, UPDATE, REFERENCES
    ON `cnpj\_%`.*
    TO 'cnpj_loader'@'%';
```

Esse grant permite operar schemas como:

```text
cnpj_loader_control
cnpj_2026_08_001
```

Ele não concede privilégios globais em `*.*`.

Se `CNPJ_LOADER_CONTROL_SCHEMA` utilizar um nome que não começa com `cnpj_`, conceda os mesmos privilégios especificamente ao schema configurado:

```sql
GRANT CREATE, SELECT, INSERT, UPDATE, REFERENCES
    ON `nome_do_schema_de_controle`.*
    TO 'cnpj_loader'@'%';
```

O host informado nos comandos `CREATE USER` e `GRANT` precisa ser o mesmo.

## Limitação dos curingas

O MySQL 8.4 ainda aceita `%` e `_` em nomes de banco utilizados por `GRANT`, mas esse comportamento está marcado como obsoleto pela documentação oficial.

Além disso, múltiplos grants com padrões sobrepostos podem produzir resultados inesperados. Por isso:

- evite grants diferentes que correspondam ao mesmo schema;
- mantenha o mesmo conjunto de privilégios para todos os schemas administrados pelo loader;
- prefira uma instância dedicada quando isso for viável;
- reavalie essa estratégia em versões futuras do MySQL.

A revisão da nomenclatura dos schemas antes da primeira release também deverá considerar a estratégia de provisionamento de privilégios.

## Verificação da conta

Conecte-se com a conta operacional e consulte os grants:

```sql
SHOW GRANTS;
```

A conta não deve possuir privilégios administrativos ou globais desnecessários.

Também é possível validar a conexão do loader executando:

```bash
(
  set -a
  source .env
  set +a

  bin/cnpj-loader migrate-control
)
```

Esse comando cria, atualiza ou valida o schema de controle sem iniciar uma carga de dados.

## Referências oficiais

- [LOAD DATA Statement — MySQL 8.4](https://dev.mysql.com/doc/refman/8.4/en/load-data.html)
- [Privileges Provided by MySQL — MySQL 8.4](https://dev.mysql.com/doc/refman/8.4/en/privileges-provided.html)
- [GRANT Statement — MySQL 8.4](https://dev.mysql.com/doc/refman/8.4/en/grant.html)
