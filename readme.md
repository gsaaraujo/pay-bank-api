Nossa **startup fictícia** lançou um banco digital para transferências instantâneas entre **clientes internos**.

Requisitos do produto no início:

- Transferências básicas entre contas do mesmo banco.
- Histórico simples de transações.
- Sem integrações externas.

Os contextos vão guiar a evolução da arquitetura de acordo com as necessidades.
**Necessidades obrigatórias: Testes de integração, pipeline de CI/CD, métricas, logs e Trunk-Based Development como estratégia de versionamento.**

# Contexto 1

A ideia já é validada mas não sabemos se as pessoas vão aderir ao nosso produto.
Então a princípio temos **zero usuários**.

### Infra inicial:

- ECS Fargate
- ALB + ACM
- RDS (postgres)

---

In progress
