# Estudos de índices

Este diretório registra os experimentos utilizados para definir os índices dos schemas de dados.

## Princípio

Os índices serão derivados de consultas reais ou representativas dos consumidores. Não criaremos índices apenas porque uma coluna parece importante.

As tabelas serão carregadas inicialmente sem chaves e sem índices. A construção ocorrerá depois da carga dos dados.

## Avaliações previstas

Os estudos deverão comparar:

- busca exata pelo CNPJ completo;
- índice de coluna única;
- índice composto por raiz, sequência da unidade e dígitos verificadores;
- chave primária e índice comum;
- índices necessários para relacionamentos;
- índices para filtros e paginação;
- criação agrupada ou separada;
- criação sequencial ou paralela em tabelas diferentes.

## Evidências

Para cada índice candidato, registraremos:

- consulta atendida;
- tempo de criação;
- espaço ocupado;
- plano de execução;
- latência antes e depois;
- linhas examinadas;
- custo para a carga e para a manutenção do schema.

Um índice somente será mantido quando estiver relacionado a uma necessidade concreta e apresentar benefício mensurável.