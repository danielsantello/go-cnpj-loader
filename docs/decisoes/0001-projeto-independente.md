# ADR 0001 — Projeto independente de sistemas consumidores

- Estado: `accepted`
- Data: 2026-08-25

## Contexto

Os dados públicos de CNPJ disponibilizados pela Receita Federal podem ser utilizados por diferentes aplicações, serviços e ambientes.

A rotina responsável por obter e carregar esses dados possui responsabilidade própria, ciclo de execução independente e requisitos operacionais específicos para o processamento de grandes volumes.

O projeto também será utilizado como estudo de Go, MySQL e engenharia de dados, além de compor um portfólio público.

Acoplar o loader a uma API, produto ou infraestrutura consumidora reduziria sua capacidade de reutilização e misturaria responsabilidades distintas.

## Decisão

O projeto será desenvolvido como aplicação independente no repositório público `go-cnpj-loader`.

O executável será chamado `cnpj-loader`.

O programa será responsável por:

- identificar publicações dos dados públicos de CNPJ;
- baixar e preparar os arquivos;
- carregar os dados em schemas versionados no MySQL;
- criar os índices definidos para consulta;
- registrar informações de controle, execução e diagnóstico;
- excluir versões somente mediante comando explícito.

O programa:

- não dependerá de uma API consumidora;
- não conhecerá a configuração dos sistemas que consultam os dados;
- não alterará automaticamente qual schema um consumidor utiliza;
- receberá por configuração os dados de conexão e os caminhos operacionais;
- poderá ser utilizado em qualquer ambiente MySQL compatível;
- manterá documentação própria;
- utilizará português na documentação e nas mensagens destinadas ao usuário;
- utilizará inglês no código, nos comandos, nos parâmetros e nos identificadores técnicos.

## Consequências

### Positivas

- responsabilidade mais clara;
- possibilidade de reutilização;
- menor acoplamento;
- melhor apresentação como projeto de portfólio;
- arquitetura e operação documentadas de forma autônoma;
- sistemas consumidores podem evoluir independentemente.

### Negativas

- configurações de integração precisam ser explícitas;
- cada consumidor é responsável por apontar para o schema desejado;
- mudanças no ambiente consumidor não podem ser resolvidas automaticamente pelo loader;
- documentações específicas de integração devem permanecer fora deste repositório.

## Condições para revisão

Esta decisão poderá ser revista somente se surgir uma necessidade técnica comprovada que altere a responsabilidade central do produto.

A simples existência de um sistema consumidor não será suficiente para justificar acoplamento.
