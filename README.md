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

**O Contexto: Do Sucesso Tático à Expansão Comercial**  
A infraestrutura de controle de concorrência distribuída desenvolvida na fase anterior (Problema 2) provou ser um sucesso tático inquestionável. A frota de drones autônomos obteve êxito no monitoramento contínuo das rotas marítimas no Estreito de Ormuz, garantindo a exclusão mútua e impedindo que múltiplos drones fossem alocados para a mesma anomalia. Graças a essa estabilidade, a operação expandiu e passou a ser financiada por um vasto consórcio internacional, composto por diferentes nações e dezenas de companhias de navegação comercial que dependem do estreito para o escoamento global.

**O Problema: A Crise de Confiança e o Cenário Bizantino**  
Com a entrada de capital financeiro e o aumento drástico das tensões geopolíticas na região, o paradigma da rede mudou. O sistema, que antes lidava apenas com falhas acidentais de conexão (*Crash-Failures*), passou a enfrentar um ambiente hostil de desconfiança mútua e espionagem industrial. Duas graves vulnerabilidades de transparência foram detectadas na arquitetura anterior:

1. **Adulteração de Provas (Integridade):**   Atores maliciosos interceptavam a rede para modificar os laudos gerados pelos drones, ocultando atividades ilícitas ou forjando rotas seguras para prejudicar nações rivais.
2. **Fraude Financeira (Duplo Gasto):** Companhias de navegação começaram a explorar a latência natural do sistema assíncrono para solicitar múltiplas escoltas de drones simultaneamente, utilizando o mesmo fundo de créditos antes que a rede pudesse sincronizar os saldos.

**A Solução: Transição para uma Arquitetura *Zero Trust***  
Para restaurar a ordem no Estreito de Ormuz, a confiança não poderia mais ser presumida; ela precisava ser matematicamente provada. O requisito fundamental para a nova arquitetura era erradicar qualquer ponto central de controle (como um "Banco Central" ou servidor mestre), uma vez que qualquer entidade centralizadora seria um alvo imediato de corrupção ou ataques cibernéticos.

Dessa forma, o sistema foi inteiramente reescrito sob o paradigma **Trustless (Zero Trust)**. A infraestrutura agora exige que todas as transações de créditos sejam seladas com **Criptografia Assimétrica (ECDSA)** pelas próprias companhias de navegação. Além disso, o estado global do sistema deixou de ser volátil; todos os despachos de drones e os seus respectivos laudos são agora eternizados em uma **Blockchain Nativa (Ledger Distribuído)**, mantida e validada através de consenso contínuo por todos os nós P2P da malha.

---

## 🎯 Objetivo do Projeto

O objetivo central deste projeto é desenvolver e consolidar uma infraestrutura distribuída operando sob o paradigma **Zero Trust (Confiança Zero)**, capaz de orquestrar a logística de drones autônomos e gerir um consórcio financeiro internacional sem a necessidade de qualquer autoridade centralizadora.

Para responder à crise de confiança no Estreito de Ormuz, o ecossistema foi projetado para assegurar quatro pilares fundamentais:

1. **Transparência Económica e Autonomia:** Implementar um motor financeiro descentralizado (*Ledger*) onde as companhias de navegação transacionam créditos para o financiamento de escoltas marítimas. O sistema deve auditar saldos de forma distribuída, deduzir custos operacionais na *Mempool* e injetar subsídios de forma totalmente autônoma através de Eleição Dinâmica de Líder, eliminando definitivamente a figura de um "Banco Central".
2. **Segurança Criptográfica e Resiliência a Fraudes:** Proteger a malha de comunicação contra atores maliciosos (Falhas Bizantinas). O sistema exige o uso de Criptografia de Curva Elíptica (ECDSA P-256) para garantir a identidade dos emissores e a integridade semântica dos pacotes, mitigando ativamente vetores de ataque cibernético como o Duplo Gasto (*Double Spending*), Ataques de Repetição (*Replay Attacks*) e a adulteração de coordenadas geográficas em trânsito.
3. **Auditoria e Imutabilidade (Blockchain):** Projetar e executar uma *Blockchain* nativa, estruturada através do encadeamento estrito de provas criptográficas (SHA-256). A meta é garantir que todas as transações financeiras e os laudos gerados pelos drones (ex: "rota segura" ou "anomalia detetada") sejam cravados num registo perpétuo, unificado pelo consórcio e matematicamente à prova de modificações póstumas.
4. **Consenso Híbrido e Tolerância a Falhas:** Assegurar a continuidade ininterrupta do serviço através de algoritmos de consenso clássico (Ricart-Agrawala com Relógios de Lamport) para a garantia de Exclusão Mútua no voo dos drones, combinados com mecanismos de auto-cura (*Active Ping / Health Checks*). A infraestrutura deve sobreviver a mortes silenciosas de servidores (*Crash-Failures*) e manter o estado da rede sincronizado.

---

## 🏗️ Arquitetura do Sistema: Evolução do Problema 2 para o Problema 3 (Relatório Técnico)

No **Problema 2**, a infraestrutura operava sob um modelo de **confiança implícita** (*Crash-Tolerant*). Os nós partilhavam uma fila distribuída em memória e resolviam a exclusão mútua para o despacho de drones através do algoritmo de **Ricart-Agrawala**. Contudo, essa arquitetura apresentava uma vulnerabilidade crítica num cenário de **Falhas Bizantinas**: se um nó fosse comprometido e propagasse mensagens forjadas (ex: adulterando o status de um drone), a malha distribuída aceitaria a instrução cegamente.

Neste **Problema 3**, para responder à premissa de um consórcio não-confiável, a arquitetura passou por uma reestruturação profunda suportada pelo modelo **Zero Trust** (Confiança Zero), resultando nas seguintes evoluções e *trade-offs*:

* 💾 **Do Estado Implícito para o Estado Derivado (Ledger):** O estado global da rede (saldos financeiros de cada companhia e histórico de missões) deixou de ser mantido "solto" em variáveis voláteis na memória (RAM). Agora, o estado é estritamente **derivado da leitura de um Livro-Razão Imutável (Blockchain)** persistido no disco físico de cada nó. Quedas de energia ou reinicializações já não causam amnésia no sistema.
* 🛡️ **Auditoria na Borda (Padrão Fail-Fast):** A lógica de ingestão de missões foi isolada atrás de um rigoroso *middleware* de validação. Requisições anómalas, com fundos insuficientes ou tentativas de *Replay Attack* são intercetadas e descartadas imediatamente na porta TCP do Broker recetor. Ao barrar a fraude na "fronteira", o sistema evita que o pacote malicioso engatilhe o protocolo de Ricart-Agrawala, poupando um enorme *overhead* (largura de banda) de comunicação na malha P2P.
* ⚖️ **Arquitetura de Consenso Híbrido:** O sistema evoluiu para operar com dois motores de consenso em paralelo. Mantivemos o **Ricart-Agrawala** (com Relógios Lógicos de Lamport) para garantir a exclusão mútua física no despacho dos drones. Paralelamente, implementámos o algoritmo de **Eleição Dinâmica de Líder** para governar a emissão de subsídios do Consórcio (o "Banco Central"), prevenindo colisões de blocos concorrentes (*Forks*) sem a necessidade de introduzir uma Prova de Trabalho (PoW) custosa.
* ⏱️ **Trade-offs Arquiteturais (Segurança vs. Desempenho):** A transição para um ambiente *Trustless* não é gratuita. O aumento colossal na segurança introduziu um custo direto no desempenho e na latência. O sistema exige agora um maior tempo de processamento computacional para calcular e verificar as assinaturas matemáticas nas Curvas Elípticas (ECDSA P-256) a cada requisição, além de validar de forma iterativa o encadeamento de *hashes* (SHA-256) de toda a cadeia a cada novo bloco minerado.

---

## 🚀 Principais Funcionalidades (Features)

* ⛓️ **Blockchain Nativa em Go:** O sistema implementa um *Ledger* imutável desenvolvido do zero. Cada missão concluída gera um bloco selado com provas criptográficas (SHA-256) que se conecta ao *hash* do bloco anterior, garantindo que o histórico de operações no Estreito de Ormuz jamais possa ser adulterado ou apagado por nós maliciosos.
* 🔐 **Assinatura Digital (ECDSA):** Autenticação rigorosa de todas as transações financeiras utilizando a Curva Elíptica P-256. O payload de cada missão (incluindo o valor pago e o setor de destino) é assinado com a chave privada do emissor, mitigando ataques de *Man-in-the-Middle*, fraudes de identidade e adulterações de rota em trânsito.
* 🏦 **Motor Económico Descentralizado:** Prevenção absoluta contra o ataque de Duplo Gasto (*Double Spending*). O sistema não armazena saldos em variáveis voláteis; ele calcula o patrimônio em tempo real cruzando os gastos confirmados no Livro-Razão com os fundos congelados (*cativo*) nas missões que ainda estão em andamento na fila ativa (*Mempool*).
* 👑 **Eleição Dinâmica de Líder:** Eliminação de pontos únicos de falha na emissão de subsídios do consórcio. A rede P2P elege de forma autônoma o nó ativo com o menor ID para atuar como emissor de moeda. Se este líder sofrer uma queda (*Crash*), a malha atualiza a topologia e transfere a liderança instantaneamente para o próximo nó elegível.
* 🛡️ **Tolerância a Falhas (Catch-up e Warm-up):** Mecanismo de auto-cura para recuperação de estado (*Crash Recovery*). Quando um Broker cai e é reinicializado, ele entra em modo de *Warm-up*, interroga os vizinhos sobre o estado da rede, descarrega o *delta* de blocos ausentes (*Catch-up*) e sincroniza o seu disco local para alcançar a malha global perfeitamente.
* 💾 **Persistência de Identidade (Wallet Recovery):** O sistema implementa uma mecânica de recuperação automática de carteiras locais. Se um cliente iniciar o terminal com o mesmo nome de companhia na mesma máquina, o sistema varre o diretório (`/data`), reconhece e carrega as chaves criptográficas (`.pem`) pré-existentes. Isso garante a retenção da conta e do histórico entre sessões, simulando perfeitamente o comportamento de uma *Crypto Wallet* real.
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
Quando uma Companhia de Navegação (Cliente) submete uma nova missão, o pacote TCP não entra na rede de imediato. Ele passa por um rigoroso *middleware* de validação tripla na "fronteira" do Broker recetor (*Fail-Fast*):

* 🔐 **Auditoria de Integridade (ECDSA):** O Broker extrai a `TransacaoToken`, regenera o *Hash* cruzando o valor financeiro com a Rota (Setor) solicitada, e valida a Assinatura Digital na Curva Elíptica P-256. Isso garante que a missão não foi interceptada nem forjada em trânsito.
* 🏦 **Auditoria Financeira (Anti-Double Spend):** O Broker recalcula dinamicamente o patrimônio da companhia através da fórmula: `Saldo = (Blocos da Blockchain + Subsídios) - Mempool`. Se não houver fundos suficientes livres, o pacote é sumariamente descartado.
* 🛡️ **Auditoria de Unicidade (Anti-Replay Attack):** O Broker rastreia a memória volátil (Fila Ativa) e o histórico persistido (Blockchain) em busca do `ID` da missão. Se o pacote for uma duplicata maliciosa reenviada por um espião, a invasão é bloqueada.

### 2. Consenso e Criação do Bloco (A Regra do "Sponsor")
Para evitar colisões de blocos (*Forks*) sem o imenso gasto energético de uma Prova de Trabalho (PoW), o sistema adota a **Regra de Autoria** para a escrita do *Ledger*:

* 🏁 **Conclusão e Laudo:** Quando um drone conclui o voo, todos os Brokers atualizam o status localmente. Contudo, **apenas o Broker Originador (Sponsor)** da missão adquire a autoridade do consórcio para redigir o Laudo de Auditoria.
* 🧱 **Selagem Criptográfica:** O *Sponsor* empacota a Missão, a Transação original e o `HashAnterior` (elo rigoroso com a cadeia), processando a prova de integridade em SHA-256 para gerar o `HashAtual` (Selo do Bloco).
* 📡 **Broadcast e Validação:** O Bloco recém-minerado é injetado localmente e propagado para a malha. Os restantes Brokers não aceitam o bloco cegamente; eles recalculam o Hash e atestam o encadeamento perfeito antes de anexarem a verdade imutável aos seus discos rígidos.

### 3. Tolerância a Falhas e Recuperação de Estado (Crash Recovery)
Em um ambiente hostil de operação, os nós da malha P2P estão sujeitos a quedas abruptas de energia ou destruição física (*Crash-Failures*). O sistema garante a resiliência global em três camadas de redundância:

* 🔄 **Recuperação do Livro-Razão (Catch-up):** Quando um nó cai e é reiniciado, ele aciona a rotina de *Warm-up*. O nó interroga a malha sobre o tamanho das *Blockchains* ativas. Se o seu disco local estiver desatualizado, ele solicita o *download* do *delta* ausente e sobrescreve o seu ficheiro `.json` para alcançar a sincronia de estado perfeita antes de voltar a operar.
* 🛟 **Adoção de Missões Órfãs (Failover de Mempool):** Se o *Broker A* for destruído enquanto coordenava a alocação de um drone, a missão ficaria congelada no limbo. Através do `processoAdocaoOrfaos`, ao detetar a morte do *Broker A* via *Active Ping*, o nó sobrevivente com o menor ID herda a autoria da missão e dá continuidade automática ao fluxo de despacho.
* 🏛️ **Sucessão do Banco Central (Failover de Liderança):** A injeção autônoma de subsídios do Consórcio é orquestrada exclusivamente pelo nó ativo com o menor ID (Líder Eleito). Caso este líder sofra um *crash*, a tabela de roteamento expurga o seu IP e, no ciclo temporal seguinte, o próximo nó elegível assume instantaneamente o dever de emitir os créditos, prevenindo a paralisia da economia distribuída.

---

## 🛡️ Matriz de Ameaças e Defesas Implementadas

O sistema foi submetido a rigorosos testes de Engenharia do Caos e Segurança Criptográfica. A tabela abaixo mapeia os vetores de ataque mitigados pela nova arquitetura *Zero Trust*:

| Ameaça / Vetor de Ataque | Mecanismo de Defesa Implementado | Efeito Prático na Arquitetura e Código |
| :--- | :--- | :--- |
| 🎭 **Falsidade Ideológica (*Spoofing*)** | Criptografia Assimétrica (ECDSA P-256) | Validação estrita de chaves públicas (`.pem`) na borda da rede. Pacotes com assinaturas forjadas são imediatamente descartados no *Socket TCP*, aplicando o padrão *Fail-Fast*. |
| 🔀 **Adulteração de Rota (*Man-in-the-Middle*)** | *Hash* Criptográfico Multivariável | O *payload* assinado concatena o valor financeiro com a coordenada geográfica (Setor). Qualquer mutação em trânsito invalida a decodificação da curva elíptica, rejeitando a sabotagem. |
| 💸 **Duplo Gasto (*Double Spending*)** | Cálculo de Estado Derivado e *Mempool* | O motor financeiro audita o histórico imutável do *Ledger* e deduz os valores retidos (*cativos*) nas missões ativas da Fila. Submissões concorrentes sem fundos disparam alertas de fraude. |
| 🔁 **Ataque de Repetição (*Replay Attack*)** | Auditoria de Unicidade de Transação | O Broker rastreia os `IDs` globais cruzando a memória volátil com o disco rígido. A injeção de cópias idênticas de um pacote legítimo esbarra na prova de existência, sendo bloqueada. |
| 👻 **Queda Silenciosa (*Silent Crash*)** | Monitorização Ativa (*Active Health Checks*) | A rotina `processoPingAtivo` varre a topologia periodicamente. Nós inativos ou em curto-circuito são expurgados, garantindo que o quórum de consenso e a Eleição de Líder nunca entrem em *Deadlock*. |
---

## 🗂️ Estrutura do Projeto

A organização do repositório segue estritamente o **Standard Go Project Layout**, isolando os pontos de entrada dos microsserviços (`/cmd`) das regras de domínio e contratos da aplicação (`/internal`).

```text
📦 economia_e_auditoria_de_guerra
 ┣ 📂 cmd
 ┃ ┣ 📂 broker             # Core do sistema: Orquestração P2P, Consenso e Blockchain
 ┃ ┣ 📂 cliente            # Interface de Linha de Comando (CLI) e Crypto Wallet
 ┃ ┣ 📂 drone              # Worker autônomo (Telemetria, Heartbeat e Execução)
 ┃ ┗ 📂 testes             # Injetor de falhas e Suíte de Integração Automatizada
 ┣ 📂 internal
 ┃ ┗ 📂 models             # Entidades de Domínio (Structs, Envelopes, Hash e Curvas ECDSA)
 ┣ 📂 data                 # Volume Persistente (Ledgers .json e Chaves Privadas .pem)
 ┣ 📜 .env                 # Mapeamento Estático de Topologia (IPs e Portas da Malha)
 ┣ 📜 docker-compose.yml   # Orquestrador da infraestrutura (Contêineres e Redes)
 ┗ 📜 README.md            # Documentação técnica, arquitetural e guia de execução

```

---

## 🛠️ Configuração de Rede (`.env`)

Por ser um sistema verdadeiramente distribuído, ele pode rodar em **uma única máquina** ou em **vários computadores físicos em um laboratório**. Para isso, é crucial configurar corretamente o arquivo `.env` na raiz do projeto:

```env
# IP do computador atual rodando este container
IP_MAQUINA=172.16.201.10

# Endereço "Semente" (Seed) de entrada na malha P2P
SEED_ADDR=172.16.201.10:9000

# Mapa estático de afinidade de nós e failover da malha
ADDR_BROKER_1=172.16.201.10:9000
ADDR_BROKER_2=172.16.201.11:9001
ADDR_BROKER_3=172.16.201.12:9002
ADDR_BROKER_4=172.16.201.13:9003

```
Para uso local, o arquivo `.env` pode ser configurado da seguinte forma, na raiz do projeto:

```env
SEED_ADDR=broker-1:9000
ADDR_BROKER_1=broker-1:9000
ADDR_BROKER_2=broker-2:9000
ADDR_BROKER_3=broker-3:9000
ADDR_BROKER_4=broker-4:9000

```
*Obs.: Necessário login no Docker Desktop*

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

# Para limpar as pastas protegidas (Linux):
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

### 3. Acompanhar os Logs

Para visualizar a topologia a ser formada, os blocos de subsídio sendo gerados nativamente pelo Líder eleito e a defesa contra ataques a funcionar em tempo real:

```bash
docker logs -f broker-1

```

*Dica: Pode usar `docker logs -f broker-X` noutro terminal para comprovar a simetria de informações na rede.*

Para visulizar o comportamento dos drones:

```bash
docker logs -f drone-alpha drone-bravo drone-beta

```

## 💻 Terminal de Comando (CLI do Cliente e Wallet)

O Cliente Interativo atua como o portal de acesso da Companhia de Navegação à rede P2P e funciona como uma verdadeira *Crypto Wallet*. Ele não armazena saldos, mas detém a responsabilidade de guardar a **Chave Privada** do utilizador e assinar digitalmente o fluxo financeiro na borda da rede.

Para iniciar a interface, execute:
```bash
docker compose run --rm --build cliente

```

**💾 Mecânica de Login e Recuperação (Wallet Recovery):**
Ao iniciar, a CLI solicita o nome da companhia. O sistema varre o diretório seguro (`/data`) buscando por um par de chaves assimétricas associado a esse nome (ex: `carteira_msc_cruzeiros.pem`). Se o ficheiro existir, **a identidade da empresa é imediatamente recuperada e autenticada**. Caso seja um novo acesso, uma nova matriz de chaves na curva ECDSA P-256 é gerada e guardada no disco.

**Painel de Operações e Auditoria:**
Após estabelecer conexão com um nó *Gateway* primário, o operador tem acesso às seguintes funcionalidades:

* 🚀 **`1` ou `2`: Submissão de Missões (Assinatura ECDSA)** - Geração de *payloads* geográficos (Manuais ou Aleatórios), assinatura da transação financeira com a chave privada local e roteamento para a *Mempool* da malha.
* ⚖️ **`5`: Auditoria de Consenso Global (Raio-X)** - Consulta o saldo efetivo da carteira interrogando **todos** os nós da rede simultaneamente. Serve para atestar matematicamente a simetria do sistema e provar a ausência de *Forks* ou *Split-Brain*.
* ⛓️ **`6`: Extração do Ledger (Blockchain)** - Efetua o *download* completo do Livro-Razão persistido no nó *Gateway*, imprimindo na consola a cadeia de blocos validada para auditoria pública de laudos e despesas.
* ☠️ **`7`: Engenharia do Caos (Menu Hacking)** - Módulo ofensivo para testar a blindagem *Byzantine Fault Tolerance* (BFT) da rede. Permite simular a injeção de pacotes forjados, adulteração de vetores de rota (*Man-in-the-Middle*) e envio de transações duplicadas (*Replay Attack*).

---

## 🤖 Suíte de Auditoria Automatizada (Integration Tests)

Para atestar a confiabilidade do modelo *Zero Trust*, foi desenvolvida uma suíte de testes de integração independente. Este módulo atua como um injetor de tráfego e de Engenharia do Caos, forjando chaves ECDSA temporárias e atacando agressivamente as barreiras da arquitetura P2P para provar a sua invulnerabilidade.

Para iniciar a auditoria completa da malha, execute:
```bash
docker compose run --rm --build testes

```

**Cenários Validados em Tempo Real (O Script Testará):**

* 🌐 **1. Conectividade com a Malha P2P:** Valida o mapeamento de rede e a acessibilidade das portas TCP de todos os nós (*Brokers*) instanciados no *cluster*.
* 🔏 **2. Envio Criptografado Legítimo (ECDSA):** Simula o comportamento de um cliente autêntico, gera uma transação selada na Curva P-256 e atesta se o Broker audita e aceita o pacote corretamente na sua *Mempool*.
* 🗺️ **3. Defesa contra Adulteração de Rota (Ataque de Setor):** O script forja um pacote legítimo, mas altera maliciosamente a rota de destino em trânsito. O teste atesta se o Broker deteta a quebra do *Hash* de origem e rejeita a invasão na porta de entrada.
* 🔁 **4. Defesa contra Replay Attack:** Dispara intencionalmente um pacote exato duplicado contra a rede para garantir que o mecanismo de indexação do Broker bloqueia a fraude e previne o consumo infinito de recursos.
* ⚖️ **5. Consenso Global de Saldos (Prevenção de Duplo Gasto):** Efetua uma varredura simultânea ("Raio-X") interrogando todos os nós da malha independentemente. O teste exige que os valores financeiros calculados sejam idênticos em todos os Brokers, provando a ausência de bifurcações (*Split-Brain*) na economia.
* 🧱 **6. Integridade Criptográfica do Ledger:** Efetua o *download* completo da *Blockchain*, recalcula iterativamente todos os *Hashes* (SHA-256) na memória e audita a ligação ao `HashAnterior`. O teste atesta categoricamente que nenhum byte do disco rígido local sofreu manipulação póstuma.

---

## 📚 Referências e Links Úteis

**Golang & Infraestrutura Nativa**
* [Documentação Oficial do Go (Golang)](https://go.dev/doc/) — Base conceitual para o uso de *Goroutines*, canais e concorrência segura na malha.
* [Pacote `net` do Go](https://pkg.go.dev/net) — Implementação de Sockets TCP puros, `DialTimeout` e controle de *Deadlines* para a comunicação assíncrona da topologia P2P.
* [Pacote `encoding/json` do Go](https://pkg.go.dev/encoding/json) — Utilizado para a serialização e desserialização do envelope de transporte padrão (`MensagemDistribuida`).
* [Go by Example: Mutexes](https://gobyexample.com/mutexes) — Referência prática para implementação de *Thread-Safety* e prevenção de *Race Conditions* na alteração do estado global da *Mempool*.

**Algoritmos Distribuídos e Consenso**
* [Distributed Systems: Principles and Paradigms (Tanenbaum & Van Steen)](https://www.distributed-systems.net/index.php/books/ds3/) — Obra clássica e referência acadêmica principal para os conceitos de Tolerância a Falhas Bizantinas (BFT), exclusão mútua e arquiteturas *Peer-to-Peer* abordados no projeto.
* [Time, Clocks, and the Ordering of Events in a Distributed System (Leslie Lamport)](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) — Artigo seminal que embasa a ordenação causal (Relógios Lógicos de Lamport) das missões na fila do sistema.
* [Ricart-Agrawala Algorithm (Wikipedia)](https://en.wikipedia.org/wiki/Ricart%E2%80%93Agrawala_algorithm) — Visão formal do algoritmo de exclusão mútua totalmente distribuído, adotado e otimizado para decidir qual broker possui o direito de despachar um drone.

**Segurança Criptográfica e Blockchain**
* [Pacote `crypto/ecdsa` do Go](https://pkg.go.dev/crypto/ecdsa) — Base da camada *Zero Trust*. Implementação da Curva Elíptica P-256 para geração de pares de chaves criptográficas (`.pem`) e assinaturas digitais.
* [Pacote `crypto/sha256` do Go](https://pkg.go.dev/crypto/sha256) — Função matemática de dispersão (*Hash*) utilizada para o encadeamento inquebrável dos blocos e auditoria de integridade do *Ledger*.
* [Replay Attack Mitigation (GeeksforGeeks)](https://www.geeksforgeeks.org/what-is-a-replay-attack/) — Embasamento teórico para a construção da defesa contra a injeção repetida de transações financeiras.

---

## 👨‍💻 Autor

Desenvolvido por **Ítallo de Santana Guimarães**  
*Engenharia de Computação — Universidade Estadual de Feira de Santana (UEFS)*

[![LinkedIn](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/itallo-guimaraes)
[![GitHub](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/ItalloGuimaraes)

## 📄 Licença

Distribuído sob a licença **MIT**. Consulte o ficheiro [LICENSE](LICENSE) para obter mais detalhes sobre as permissões e restrições de uso.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
