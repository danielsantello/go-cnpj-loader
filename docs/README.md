# Documentação do go-cnpj-loader

Esta documentação registra a arquitetura, a operação, as decisões técnicas e os experimentos realizados durante o desenvolvimento do `go-cnpj-loader`.

O conteúdo é construído junto com o projeto. Decisões relevantes não são apresentadas apenas como conclusões: sempre que possível, preservamos o contexto, as alternativas avaliadas e as evidências utilizadas.

## Estrutura

### Arquitetura

Documentos que descrevem a responsabilidade do programa, seus componentes, seus limites e o fluxo geral de execução.

- [Visão geral](arquitetura/visao-geral.md)

### Desenvolvimento

Orientações para preparar o ambiente, executar verificações de qualidade e trabalhar no código do projeto.

- [Testes](desenvolvimento/testes.md)

### Decisões arquiteturais

Registros de Decisão Arquitetural, também chamados de ADRs.

Cada registro contém:

- contexto;
- problema;
- alternativas consideradas;
- decisão;
- consequências;
- condições para revisão.

Uma decisão substituída não é apagada. Um novo ADR registra a mudança e referencia a decisão anterior.

- [Índice de decisões](decisoes/README.md)

### Estudos

Experimentos reproduzíveis utilizados para escolher estratégias técnicas.

- [Carga de dados](estudos/carga/README.md)
- [Índices](estudos/indices/README.md)
- [Busca textual](estudos/busca-textual/README.md)

### Operação

Procedimentos de instalação, configuração, execução, diagnóstico, exclusão de versões e limpeza do workspace.

A documentação operacional será adicionada conforme os comandos forem implementados.

### Referência

Contratos técnicos, códigos de saída, configurações, schemas e outros materiais de consulta.

A documentação de referência será adicionada progressivamente.
