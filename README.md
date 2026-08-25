# go-cnpj-loader

> **Em construção.** Carregador versionado dos dados públicos de CNPJ da Receita Federal, desenvolvido em Go.

## Sobre o projeto

O `go-cnpj-loader` é uma aplicação de linha de comando responsável por obter, preparar e carregar no MySQL os dados públicos de CNPJ disponibilizados pela Receita Federal.

O projeto é independente de qualquer API consumidora. Cada carga produz um novo schema versionado, permitindo que a publicação e a remoção de versões anteriores sejam realizadas posteriormente, de forma explícita e controlada.

Além de sua finalidade operacional, o projeto também funciona como estudo de caso sobre Go, MySQL e processamento eficiente de grandes volumes de dados.

## Princípios

- a Receita Federal é considerada a fonte soberana dos dados;
- não são realizadas validações semânticas ou correções cadastrais;
- cada carga cria um novo schema;
- a publicação de uma versão é uma decisão externa e manual;
- versões anteriores nunca são removidas automaticamente;
- tabelas são carregadas antes da criação dos índices;
- estratégias de carga e indexação são escolhidas por meio de experimentos reproduzíveis;
- decisões arquiteturais e resultados de desempenho são documentados.

## Tecnologias

- Go;
- MySQL 8.4;
- Docker e Docker Compose;
- GitHub Actions.

## Estado atual

A fundação do projeto está em desenvolvimento. A arquitetura inicial foi definida e será documentada progressivamente no repositório.

## Documentação

A documentação técnica, os registros de decisões arquiteturais e os estudos de desempenho serão mantidos no diretório `docs`.

## Licença

Este projeto é distribuído sob a licença MIT.
