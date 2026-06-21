# 🚢 Economia e Auditoria de Guerra: Estreito de Ormuz

![Go Version](https://img.shields.io/badge/Go-1.21-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)
![Architecture](https://img.shields.io/badge/Arquitetura-P2P_|_Blockchain-8A2BE2?style=for-the-badge)
![Security](https://img.shields.io/badge/Segurança-ECDSA_|_SHA256-D32F2F?style=for-the-badge)
![Status](https://img.shields.io/badge/Status-Concluído-4CAF50?style=for-the-badge)

Este projeto resolve o **Problema 3** da disciplina **TEC502: MI - Concorrência e Conectividade** (Universidade Estadual de Feira de Santana - UEFS). 

O sistema implementa uma infraestrutura distribuída **Byzantine Fault Tolerant (BFT)**, que orquestra uma frota de drones marítimos utilizando uma rede *Peer-to-Peer* (P2P), Criptografia Assimétrica e uma *Blockchain* (Ledger Distribuído) construída do zero. O objetivo é garantir a total transparência, auditoria e prevenção de fraudes num ambiente de consórcio onde não existe confiança entre os intervenientes.

---

## 📖 Sobre o Projeto
Após o sucesso tático da alocação de drones (Problema 2), a operação no Estreito de Ormuz passou a ser financiada por um consórcio internacional de companhias de navegação. Com o aumento da tensão geopolítica, surgiu um cenário de desconfiança: adulteração de laudos, ataques de interceção e tentativas de desviar recursos ("duplo gasto"). 

Para restaurar a confiança sem depender de um "Banco Central" ou de um servidor mestre vulnerável, a infraestrutura foi reescrita para ser **Trustless** (Zero Trust), baseando-se em selos criptográficos e consenso distribuído absoluto.

---

## 🎯 Objetivo do Projeto
Desenvolver um ecossistema descentralizado que assegure:
1. **Transparência Económica:** Contabilização descentralizada de créditos e custos operacionais das missões.
2. **Prevenção de Fraudes:** Mitigação algorítmica e criptográfica de Duplo Gasto (*Double Spending*), Ataques de Repetição (*Replay Attacks*) e Adulteração de Rotas.
3. **Auditoria e Imutabilidade:** Registo perpétuo e inalterável dos laudos das missões através de uma *Blockchain* gerida pelos próprios nós P2P.

---

## 🏗️ Arquitetura do Sistema: Problema 2 vs Problema 3 (Relatório)

No **Problema 2**, o sistema baseava-se numa confiança implícita. Os nós partilhavam uma fila distribuída e resolviam a exclusão mútua via algoritmo de **Ricart-Agrawala**. Contudo, o sistema falhava num cenário bizantino: se um nó enviasse informações forjadas, a malha aceitava cegamente.

Neste **Problema 3**, a arquitetura sofreu uma evolução profunda (Trade-offs arquiteturais):
* **Fim do Estado Implícito:** O estado atual (Saldos e Conclusões) não é mais mantido solto na memória (RAM). Ele é **derivado de um Livro-Razão Imutável (Ledger)** persistido em disco.
* **Segurança na Fronteira (Fail-Fast):** A lógica de aceitação de missões foi enriquecida com auditorias em tempo real. Um pacote anómalo nem sequer chega a engatilhar o Ricart-Agrawala, poupando a largura de banda da rede.
* **Consenso Híbrido:** Mantivemos o *Ricart-Agrawala* com Relógios Lógicos de Lamport para o **despacho físico dos drones** (Exclusão Mútua), mas implementámos um algoritmo de **Eleição Dinâmica de Líder** para a injeção governamental de créditos (Subsídios), evitando colisões de mineração (*Forks*).
* **O Custo (Trade-off):** O aumento formidável na segurança exige maior tempo de processamento computacional para a verificação das Curvas Elípticas (ECDSA) e a validação do encadeamento de *hashes*, adicionando latência em prol da integridade absoluta.

---

## 🚀 Principais Funcionalidades (Features)

* ⛓️ **Blockchain Nativa em Go:** Encadeamento inquebrável de laudos através do algoritmo SHA-256.
* 🔐 **Assinatura Digital (ECDSA):** Autenticação de pacotes em trânsito com a Curva P-256, impedindo a fraude de identidade e ataques *Man-in-the-Middle*.
* 🏦 **Motor Económico Descentralizado:** Cálculo de saldo *On-the-fly* integrando o Livro-Razão consolidado com o saldo cativo na *Mempool* (Fila Ativa).
* 👑 **Eleição Dinâmica de Líder:** Sucessão automática de responsabilidades (Banco Central temporário) caso o nó autoritário seja destruído.
* 🛡️ **Catch-up e Warm-up:** Nós que caem e regressam interrogam a rede, descarregam o *delta* de blocos faltantes e atualizam os seus ficheiros locais (Recuperação de estado).

---

## 📡 Especificação do Protocolo (API de Comunicação)

Toda a comunicação na malha P2P é assíncrona, efetuada via *Sockets TCP*, utilizando o formato padronizado (Envelope Universal). O roteamento interno e a máquina de estados do Broker reagem de acordo com o `Tipo` da mensagem recebida:

```json
{
  "tipo": "SYNC_NEW",
  "sender_id": 1,
  "timestamp": 1781833769,
  "payload": { ... }
}

```

### Dicionário de Mensagens do Sistema

| Tipo (String) | Constante Interna | Payload Esperado (`interface{}`) | Descrição / Ação do Broker |
| --- | --- | --- | --- |
| `JOIN` | `MsgJoin` | `JoinRequest` | *Handshake* inicial para entrar na malha P2P. O broker valida e adiciona o IP remetente à sua tabela de roteamento (*Peers*). |
| `JOIN_ACK` | `MsgJoinACK` | `JoinRequest` | Confirmação bidirecional do *handshake*, consolidando a topologia de rede descentralizada. |
| `SYNC_NEW` | `MsgSyncNew` | `Requisicao` | Ingestão de novo alerta. O Broker submete o pacote à **Auditoria ECDSA, Auditoria de Saldo e Prevenção de *Replay Attack*** antes de o aceitar na *Mempool*. |
| `SYNC_UPDATE` | `MsgSyncUpdate` | `Requisicao` | Atualização atômica do status de uma missão global (ex: transição de `PENDENTE` para `EM_ATENDIMENTO`). |
| `FULL_SYNC` | `MsgFullSync` | `[]Requisicao` | Transferência completa de estado (*Snapshot* da *Mempool*) para nós recém-conectados na rede (Recuperação de Memória). |
| `REQ_DRONE` | `MsgReqDrone` | `string` (ID) | Disparo do algoritmo **Ricart-Agrawala** exigindo permissão da rede, ou *Polling* direto de um Drone procurando trabalho. |
| `REPLY_OK` | `MsgReplyOK` | `string` (ID) | Voto de aprovação no algoritmo de consenso distribuído. A missão só avança quando o quórum for atingido. |
| `DRONE_HEARTBEAT` | `MsgDroneHeartbeat` | `DroneStatus` | Sinal de vida (*Keep-Alive*) do drone em voo. Interrompe a contagem do *Watchdog* de falhas do Broker. |
| `DRONE_CONCLUIDO` | `MsgDroneConcluido` | `DroneStatus` | O drone sinaliza o término físico. Autoriza o Broker *Sponsor* (Originador) a redigir o Laudo e selar o Bloco na Blockchain. |
| `CONSULTA_FILA` | `MsgConsultaFila` | *(Vazio)* | Requisição passiva via Terminal do Cliente. O Broker devolve o estado imaculado da sua *Mempool* local. |
| `NOVO_BLOCO` | `MsgNovoBloco` | `Bloco` | Recebimento de um bloco recém-minerado por um *Peer*. Aciona a auditoria criptográfica rigorosa (`HashAtual` vs `HashAnterior`). |
| `CONSULTA_SALDO` | `MsgConsultaSaldo` | `string` (PubKey) | Aciona a varredura contra **Duplo Gasto**. O Broker cruza todas as faturas do *Ledger* com o dinheiro congelado na *Mempool*. |
| `CONSULTA_LEDGER` | `MsgConsultaLedger` | *(Vazio)* | Pedido de *Download* de toda a *Blockchain* Imutável. Usado para Clientes (Auditoria) e Brokers desatualizados (*Catch-up*). |

---

## 🔍 Modelagem e Fluxos do Sistema

### 1. Fluxo de Submissão e Validação (O "Triple-Check")

Quando uma Companhia (Cliente) submete uma missão:

1. **Auditoria de Integridade:** O Broker interceta a `TransacaoToken`, regenera o Hash associando a Rota (Setor) ao valor pago, e valida a Assinatura ECDSA.
2. **Auditoria Financeira:** O Broker calcula `Saldo = Blocos da Blockchain + Subsídios - Mempool`. Se for insuficiente, o pacote é descartado.
3. **Auditoria de Repetição (Anti-Replay):** O Broker varre a memória e o histórico para garantir que aquele `ID` é único, mitigando ataques de *Replay*.

### 2. Criação do Bloco (O Sponsor)

Quando um drone completa o voo:

* Apenas o **Broker Originador (Sponsor)** da missão ganha o direito de redigir o Laudo.
* Ele empacota a Missão, a Transação e o `HashAnterior`, processando o *Selo Criptográfico* (HashAtual).
* O Bloco é injetado localmente e emitido um *Broadcast*. Os restantes Brokers efetuam o processo de aceitação ou recusam caso os hashes não coincidam.

---

## 🛡️ Resumo das Defesas Implementadas

| Ameaça / Ataque | Defesa Implementada no Sistema | Efeito no Código |
| --- | --- | --- |
| **Falsidade Ideológica** | Curvas Elípticas (ECDSA) | O sistema gera e lê ficheiros `.pem`. Assinaturas inválidas são barradas na porta do socket. |
| **Ataque à Rota (Man-in-the-Middle)** | Hash Multivariável | O Setor destino é englobado na assinatura matemática. Adulterar o setor quebra o selo financeiro. |
| **Duplo Gasto (Double Spend)** | Varredura de Fila (Mempool) | O dinheiro de uma missão engatilhada fica cativo. Submissões subsequentes falham por "Saldo Insuficiente". |
| **Ataque de Repetição (Replay)** | Indexação de Contexto Global | Identificadores intersetam com a fila e o Ledger. O pacote replicado bate no muro da unicidade. |
| **Queda Silenciosa (Silent Crash)** | Monitorização Ativa (Ping) | Varreduras constantes (`processoPingAtivo`) expurgam nós inativos, protegendo a Eleição de Líder de congelamentos. |

---

## 🗂️ Estrutura do Projeto

```text
📦 economia_e_auditoria_de_guerra
 ┣ 📂 cmd
 ┃ ┣ 📂 broker      # (O Motor de Regras, P2P, Blockchain, Agrawala)
 ┃ ┣ 📂 cliente     # (O CLI da Companhia de Navegação + ECDSA Generator)
 ┃ ┣ 📂 drone       # (O Worker autônomo que executa o "Sleep" da missão)
 ┃ ┗ 📂 testes      # (A Suíte de Auditoria Ofensiva Automatizada)
 ┣ 📂 internal
 ┃ ┗ 📂 models      # (Structs: Blockchain, Messages, Requisicao, Transacao)
 ┣ 📂 data          # (Volume: Onde são armazenados os .json e .pem)
 ┣ 📜 .env          # (Mapeamento da Topologia de Rede)
 ┣ 📜 docker-compose.yml
 ┗ 📜 README.md

```

---

## 🛠️ Configuração de Rede (`.env`)

Por ser um sistema verdadeiramente distribuído, ele pode rodar em **uma única máquina** ou em **vários computadores físicos em um laboratório**. Para isso, é crucial configurar corretamente o arquivo `.env` na raiz do projeto:

```env
# IP do computador atual rodando este container
IP_MAQUINA=172.16.201.12

# Endereço "Semente" (Seed) de entrada na malha P2P
SEED_ADDR=172.16.201.11:9000

# Mapa estático de afinidade de nós e failover da malha
ADDR_BROKER_1=172.16.201.11:9000
ADDR_BROKER_2=172.16.201.12:9001
ADDR_BROKER_3=172.16.201.10:9002
ADDR_BROKER_4=172.16.201.10:9003

```
Para uso local, o arquivo `.env` deve ser configurado da seguinte forma, na raiz do projeto:

```env
SEED_ADDR=broker-1:9000
ADDR_BROKER_1=broker-1:9000
ADDR_BROKER_2=broker-2:9000
ADDR_BROKER_3=broker-3:9000
ADDR_BROKER_4=broker-4:9000

```

## ⚙️ Configuração e Execução (Como Executar)

### Pré-requisitos
* **Docker** e **Docker Compose** instalados.

### 1. Limpeza do Ambiente (A Tábula Rasa)
É altamente recomendada a limpeza da pasta de dados antes de iniciar a simulação no laboratório, garantindo a ausência de *Forks* (conflitos de blocos e carteiras de execuções anteriores):

```bash
docker compose down
# Em ambiente Linux/WSL/Laboratório:
rm -f data/*.json data/*.pem
# Obs.: É possível que haja a necessidade do uso do "sudo"

# Para limpar as pastas protegidas:
docker run --rm -v "$PWD/data:/dados" alpine sh -c "rm -f /dados/*.json /dados/*.pem"


```

### 2. Iniciar a Malha Distribuída (Warm-up e Descoberta)

**⚠️ IMPORTANTE - Subida Gradual:** Como a arquitetura P2P depende de *Handshakes* síncronos e assíncronos (`JOIN` e `JOIN_ACK`) para formar a topologia em malha, é **fortemente indicado subir os componentes aos poucos**. Se todos os contêineres forem iniciados no exato mesmo milissegundo, podem ocorrer falhas de *timeout* antes que as portas TCP do SO hospedeiro estejam totalmente abertas para escuta.

Para uma sincronização perfeita, siga esta ordem de subida:

```bash
# 1. Construa e inicie o Gateway / Nó Semente Principal
docker compose up --build -d broker1

# 2. Aguarde de 3 a 5 segundos e suba o resto do Consórcio (Brokers)
docker compose up -d broker2 broker3 broker4

# 3. Aguarde a malha estabilizar e suba a frota de Drones Trabalhadores
docker compose up -d drone_alpha drone_bravo drone_beta

```

### 3. Acompanhar os Logs do Banco Central / Líder

Para visualizar a topologia a ser formada, os blocos de subsídio sendo gerados nativamente pelo Líder eleito e a defesa contra ataques a funcionar em tempo real:

```bash
docker logs -f broker-1

```

*(Dica: Pode usar `docker logs -f broker-2` noutro terminal para comprovar a simetria de informações na rede).*

## 💻 Terminal de Comando (Cliente Operador)

O Cliente Interativo é o portal da Companhia de Navegação. Ele detém as chaves criptográficas para a assinatura digital do fluxo financeiro:

```bash
docker compose run --rm --build cliente

```

**Funcionalidades do Menu:**

* `1` ou `2`: Gerar e pagar missões (Manuais ou Aleatórias).
* `5`: **Auditoria de Consenso (Raio-X)** - Varre toda a malha perguntando pelo Saldo.
* `6`: Descarrega e imprime a cópia atualizada da *Blockchain* com os laudos inalteráveis.
* `7`: **Menu de Ataques Hacking** (Teste a robustez da rede com ataques de integridade, adulteração de rota e *Replay Attack*).

---

## 🤖 Suíte de Auditoria Automatizada (Integration Tests)

Foi construída uma suite independente que forja assinaturas e testa agressivamente as barreiras da arquitetura P2P.

```bash
docker compose run --rm --build testes

```

O script testará:

1. Conectividade com a malha P2P.
2. Envio Criptografado Legítimo (ECDSA).
3. Defesa contra Adulteração de Rota (Ataque de Setor).
4. Defesa contra Replay Attack.
5. Consenso Global de Saldos (Duplo Gasto).
6. Integridade Criptográfica do Ledger (Recálculo de Hashes de toda a cadeia).

---

## 👨‍💻 Autor

Desenvolvido por **Ítallo de Santana Guimarães** [LinkedIn](www.linkedin.com/in/itallo-guimaraes) | [GitHub](https://github.com/ItalloGuimaraes)
Engenharia de Computação - Universidade Estadual de Feira de Santana (UEFS). 

## 📄 Licença

Distribuído sob a licença **MIT**. Veja o arquivo [LICENSE](LICENSE) para mais detalhes
