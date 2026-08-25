# Estudos de carga

Este diretório registra os experimentos utilizados para escolher a estratégia de importação dos dados no MySQL.

## Estratégias iniciais

Serão avaliadas:

- `LOAD DATA INFILE` sequencial;
- `LOAD DATA INFILE` com duas conexões;
- `LOAD DATA INFILE` com quatro conexões;
- MySQL Shell com dois trabalhadores;
- MySQL Shell com quatro trabalhadores;
- `LOAD DATA LOCAL INFILE` coordenado pelo Go, caso permaneça relevante.

Os testes distinguirão cargas simultâneas destinadas à mesma tabela e cargas destinadas a tabelas diferentes.

## Métricas

Cada experimento deverá registrar:

- ambiente;
- hardware;
- publicação utilizada;
- arquivos e hashes;
- versão do loader;
- versão e configuração do MySQL;
- estratégia;
- concorrência;
- duração;
- registros por segundo;
- megabytes por segundo;
- CPU;
- memória;
- leitura e escrita no armazenamento;
- espaço utilizado;
- warnings;
- comportamento diante de interrupção.

## Regra de decisão

Nenhuma estratégia será escolhida apenas por impressão subjetiva.

A decisão deverá relacionar o ganho mensurável, o consumo de recursos, a segurança, a complexidade operacional e a capacidade de diagnóstico.
