# Estudo de busca textual

A pesquisa por nomes empresariais será estudada separadamente da carga principal e dos índices relacionais do MySQL.

## Contexto

Em uma arquitetura anterior, o Manticore Search foi considerado a solução mais adequada para pesquisa textual.

Essa conclusão será reavaliada antes de qualquer nova implementação, pois ferramentas, versões, requisitos e alternativas podem ter mudado.

## Escopo futuro

O estudo deverá definir primeiro o comportamento esperado da busca, incluindo:

- pesquisa por razão social;
- pesquisa por nome fantasia;
- prefixos;
- palavras parciais;
- tolerância a variações;
- acentuação;
- ordenação por relevância;
- filtros combinados;
- atualização do índice após a publicação de um novo schema.

Depois serão comparadas soluções como:

- Manticore Search;
- recursos textuais do MySQL;
- outros mecanismos atuais considerados adequados.

## Critérios

A decisão deverá considerar:

- qualidade dos resultados;
- latência;
- tempo de indexação;
- espaço utilizado;
- complexidade operacional;
- integração com a API;
- processo de troca entre versões dos dados;
- observabilidade e recuperação.

A busca textual não faz parte da primeira versão funcional do loader.