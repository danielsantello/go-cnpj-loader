# Visão geral da arquitetura

## Responsabilidade

O `go-cnpj-loader` é uma aplicação de linha de comando responsável por obter, preparar e carregar no MySQL os dados públicos de CNPJ disponibilizados pela Receita Federal.

O programa coordena uma carga completa, registra seu histórico e produz um novo schema versionado, tecnicamente pronto para consulta.

## Limites

O loader não é responsável por:

- atender consultas de consumidores;
- alterar a configuração de uma API;
- publicar automaticamente uma versão;
- identificar qual schema uma aplicação externa utiliza;
- excluir automaticamente versões anteriores;
- corrigir ou validar semanticamente os dados da Receita;
- fornecer pesquisa textual por nomes;
- executar como serviço permanente.

A publicação de uma versão e a remoção de versões anteriores são decisões externas e manuais.

## Princípio sobre os dados

A Receita Federal é considerada a fonte soberana dos dados cadastrais.

O loader não verifica se um CNPJ, CNAE, endereço ou qualquer outro conteúdo está semanticamente correto. Se a fonte fornecer um valor incomum, ele deverá ser preservado sempre que a estrutura técnica definida pelo projeto comportar esse valor.

A validação realizada pelo loader é estritamente técnica. São exemplos:

- integridade de arquivos;
- estrutura esperada da publicação;
- compatibilidade das colunas;
- erros retornados pelo MySQL;
- warnings que indiquem truncamento ou perda causada pelo nosso schema.

## Ciclo de uma carga

Uma execução completa deverá:

1. adquirir o bloqueio operacional;
2. validar configuração e pré-requisitos;
3. identificar a publicação;
4. criar o inventário dos arquivos;
5. realizar os downloads necessários;
6. validar tamanho e hash dos arquivos;
7. extrair os arquivos de forma segura;
8. criar um novo schema versionado;
9. criar as tabelas sem índices;
10. carregar os dados;
11. criar os índices obrigatórios;
12. registrar o resumo técnico;
13. marcar a versão como `ready`;
14. liberar o bloqueio operacional;
15. encerrar o processo.

Uma versão `ready` está tecnicamente concluída, mas não necessariamente publicada para uma API.

## Versionamento dos schemas

Cada carga cria um novo schema. Um schema existente nunca será truncado, reutilizado ou sobrescrito silenciosamente.

Para cargas reais, a convenção inicial é `cnpj_YYYY_MM_NNN`, como `cnpj_2026_07_001`.

Para distinguir a finalidade de cada carga, serão utilizados ambientes lógicos distintos:

- desenvolvimento: `cnpj_dev_YYYY_MM_NNN`;
- benchmark: `cnpj_bench_YYYY_MM_NNN`;
- produção: `cnpj_YYYY_MM_NNN`.

O número final diferencia novas tentativas e múltiplas cargas da mesma publicação.

## Publicação e exclusão

Depois que uma versão alcançar o estado `ready`, o operador poderá configurar manualmente a API consumidora para utilizar o novo schema.

A versão anterior permanecerá disponível durante o período de observação considerado necessário.

Sua exclusão somente ocorrerá por um comando explícito, como `cnpj-loader delete-version --schema cnpj_2026_06_001`.

A limpeza dos arquivos do workspace será outra operação independente, realizada por `cnpj-loader clean-workspace`.

Excluir um schema nunca implicará apagar os arquivos, e limpar arquivos nunca implicará excluir um schema.

## Schema de controle

O schema permanente `cnpj_loader_control` armazenará o histórico operacional.

Ele registrará, entre outros elementos:

- publicações;
- versões;
- execuções;
- etapas;
- arquivos;
- cargas;
- operações de índices;
- eventos;
- configurações não sigilosas;
- migrações do próprio schema de controle.

Ao contrário dos schemas de dados durante a carga, o schema de controle possuirá chaves e índices desde sua criação.

## Estratégia de carga

O Go atuará como coordenador da execução. A importação em massa será realizada pelo MySQL, utilizando estratégias baseadas em `LOAD DATA`.

As alternativas serão comparadas por experimentos reproduzíveis, incluindo:

- `LOAD DATA INFILE` sequencial;
- cargas paralelas com duas ou quatro conexões;
- MySQL Shell com diferentes quantidades de trabalhadores;
- `LOAD DATA LOCAL INFILE`, caso permaneça relevante.

A estratégia e a concorrência padrão somente serão escolhidas depois dos benchmarks.

As tabelas serão criadas sem chaves e sem índices para favorecer a carga. Os índices obrigatórios serão construídos posteriormente.

## Transações e falhas

A carga completa não constitui uma única transação.

Cada arquivo será tratado como uma unidade transacional independente. O resultado e os warnings serão analisados antes do `COMMIT`.

Quando ocorrer uma falha impeditiva:

- novas tarefas não serão iniciadas;
- a versão será marcada como `failed`;
- o schema incompleto será preservado;
- os arquivos serão preservados;
- o histórico e os detalhes técnicos serão mantidos;
- nenhuma limpeza automática será realizada.

Comandos DDL possuem limites transacionais próprios no MySQL. Por isso, o programa registra individualmente a criação de tabelas e índices e não promete um rollback global impossível.

## Execução e concorrência

Somente uma operação mutável poderá ser executada por vez.

Os comandos `load`, `delete-version` e `migrate-control` utilizarão um bloqueio consultivo no MySQL.

A concorrência útil ocorrerá dentro de uma única execução e será controlada pelo loader.

Comandos de consulta, como `list-versions`, `show-execution` e `version`, poderão ser utilizados durante outra operação.

## Interrupção

O programa tratará `SIGINT` e `SIGTERM`.

Na primeira solicitação de interrupção, deixará de iniciar novas tarefas e tentará encerrar as operações em andamento de forma controlada.

Uma segunda solicitação poderá provocar encerramento imediato.

Execuções abandonadas não serão retomadas nem apagadas silenciosamente. Uma execução posterior reconciliará o histórico disponível com o estado físico encontrado.

## Componentes

A arquitetura inicial prevê os seguintes componentes:

- CLI;
- configuração;
- coordenação da aplicação;
- schema de controle;
- conexão com MySQL;
- descoberta da publicação;
- workspace;
- download;
- extração;
- gerenciamento de schemas;
- criação de tabelas;
- execução das cargas;
- criação dos índices;
- controle das execuções;
- observabilidade;
- informações de compilação.

Os pacotes serão criados somente quando suas responsabilidades se tornarem concretas.

## Idioma

O projeto utiliza:

- português na documentação;
- português nas mensagens operacionais destinadas ao usuário;
- inglês no código Go;
- inglês nos comandos e parâmetros da CLI;
- inglês nos schemas, tabelas e campos;
- inglês nos códigos e valores internos utilizados pelo programa;
- conteúdo da Receita preservado conforme a fonte.

Elementos técnicos convencionais gerados pelo Cobra, como `Usage`, `Flags` e `Help`, permanecem em inglês.

## Decisões experimentais pendentes

Continuam abertas até a realização dos estudos:

- estratégia definitiva de carga;
- concorrência ideal;
- tipos e tamanhos físicos finais das colunas;
- transformações técnicas durante o `LOAD DATA`;
- conjunto definitivo de índices;
- estratégia de criação dos índices;
- mecanismo futuro de busca textual por nomes.

Manticore Search será reavaliado juntamente com alternativas atuais quando o estudo de busca textual for iniciado.
