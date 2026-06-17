package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"ormuz_distribuido/internal/models"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// =========================================================
// Estruturas de Dados
// =========================================================

// Broker representa o nó P2P responsável por orquestrar requisições e drones.
// Gerencia conexões assíncronas, estado distribuído e algoritmos de consenso.
type Broker struct {
	ID                 int
	Porta              string
	MeuEndereco        string
	Relogio            int64
	ListaDistribuida   []models.Requisicao
	Blockchain         models.Blockchain // Ledger: Histórico imutável de blocos
	Peers              map[int]string
	FalhasConsecutivas map[int]int
	mu                 sync.Mutex
}

// PendenciaAgrawala representa um pedido de zona crítica via protocolo de Ricart-Agrawala.
// Mantém o estado da negociação assíncrona até que o quorum seja atingido.
type PendenciaAgrawala struct {
	MissionID      string
	TimestampLocal int64
	Respostas      map[int]bool // Mapeia os IDs dos brokers que já enviaram OK (previne duplo voto).
	TotalEsperado  int
	DroneConn      net.Conn
	DroneID        string
}

// Mapa global protegido por mutex no acesso para controlar as negociações em andamento.
var pendencias = make(map[string]*PendenciaAgrawala)

// =========================================================
// Relógio Lógico de Lamport
// =========================================================

// tickRelogio incrementa o relógio local para eventos internos.
func (b *Broker) tickRelogio() int64 {
	b.Relogio++
	return b.Relogio
}

// atualizarRelogio sincroniza o relógio lógico baseado no timestamp remoto.
// Segue a regra clássica de Lamport: max(local, remoto) + 1.
func (b *Broker) atualizarRelogio(remoteTimestamp int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if remoteTimestamp > b.Relogio {
		b.Relogio = remoteTimestamp
	}
	b.Relogio++
}

// =========================================================
// Ordenação e Determinização
// =========================================================

// ordenarLista garante ordenação determinística da fila distribuída em toda a malha.
// Critérios: Prioridade (DESC) -> Timestamp Lamport (ASC) -> Broker ID (ASC, tie-breaker).
func (b *Broker) ordenarLista() {
	sort.Slice(b.ListaDistribuida, func(i, j int) bool {
		r1, r2 := b.ListaDistribuida[i], b.ListaDistribuida[j]
		if r1.Prioridade != r2.Prioridade {
			return r1.Prioridade > r2.Prioridade
		}
		if r1.Timestamp != r2.Timestamp {
			return r1.Timestamp < r2.Timestamp
		}
		return r1.BrokerID < r2.BrokerID
	})
}

// =========================================================
// Comunicação P2P (Modelo Fan-Out Paralelo)
// =========================================================

// limiteFalhas define a tolerância a quedas consecutivas antes da remoção do nó da topologia.
const limiteFalhas = 3

// broadcast dispara mensagens assíncronas concorrentes para a lista de Peers ativa.
// Retorna a contagem de pacotes entregues e os IDs dos nós removidos devido a dead-links.
func (b *Broker) broadcast(msg models.MensagemDistribuida) (enviados int, removidos []int) {
	b.mu.Lock()
	peersCopia := make(map[int]string, len(b.Peers))
	for id, addr := range b.Peers {
		peersCopia[id] = addr
	}
	b.mu.Unlock()

	var wg sync.WaitGroup
	var envMu sync.Mutex
	removidos = []int{}

	for id, addr := range peersCopia {
		if id == b.ID {
			continue
		}
		wg.Add(1)

		go func(peerID int, peerAddr string) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", peerAddr, 2*time.Second)
			if err != nil {
				b.mu.Lock()
				b.FalhasConsecutivas[peerID]++
				falhas := b.FalhasConsecutivas[peerID]
				b.mu.Unlock()

				if falhas >= limiteFalhas {
					b.mu.Lock()
					delete(b.Peers, peerID)
					delete(b.FalhasConsecutivas, peerID)
					b.mu.Unlock()

					envMu.Lock()
					removidos = append(removidos, peerID)
					envMu.Unlock()
					fmt.Printf("[P2P] Broker %d removido após %d falhas.\n", peerID, falhas)
				}
				return
			}

			b.mu.Lock()
			b.FalhasConsecutivas[peerID] = 0
			b.mu.Unlock()

			if err := json.NewEncoder(conn).Encode(msg); err == nil {
				envMu.Lock()
				enviados++
				envMu.Unlock()
			}
			conn.Close()
		}(id, addr)
	}

	wg.Wait()
	return enviados, removidos
}

// =========================================================
// Servidor TCP e Handlers
// =========================================================

// Iniciar inicializa as threads de background (Watchdog e Aging) e entra
// no loop de aceitação de conexões de entrada.
func (b *Broker) Iniciar(seeds []string) {
	ln, err := net.Listen("tcp", b.Porta)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao abrir porta %s: %v\n", b.Porta, err)
		return
	}
	defer ln.Close()

	b.processoEnvelhecimento()
	b.processoWatchdogDrones()
	b.processoAdocaoOrfaos()
	go b.buscarAtualizacaoDeLedger(seeds)
	b.processoGeradorCreditos()

	if len(seeds) > 0 && b.MeuEndereco != "" {
		go b.processoReconexao(seeds)
	}

	fmt.Printf(">>> Broker %d escutando em %s (externo: %s)...\n", b.ID, b.Porta, b.MeuEndereco)

	for {
		conn, err := ln.Accept()
		if err == nil {
			go b.handleConnection(conn)
		}
	}
}

// handleConnection faz o parsing do payload inicial e roteia o pacote
// para a função de tratamento específica de acordo com o Tipo da Mensagem.
func (b *Broker) handleConnection(conn net.Conn) {
	var msg models.MensagemDistribuida

	decoder := json.NewDecoder(conn)
	decoder.UseNumber()

	if err := decoder.Decode(&msg); err != nil {
		conn.Close()
		return
	}

	fmt.Printf("[REDE] Tipo=%-14s | De=%d | TS=%d\n", msg.Tipo, msg.SenderID, msg.Timestamp)
	b.atualizarRelogio(msg.Timestamp)

	switch msg.Tipo {
	case models.MsgJoin:
		b.processarJoin(msg, conn)
	case models.MsgJoinACK:
		b.processarJoinACK(msg)
		conn.Close()
	case models.MsgSyncNew:
		b.processarNovoAlerta(msg)
		conn.Close()
	case models.MsgSyncUpdate:
		b.processarUpdateStatus(msg)
		conn.Close()
	case models.MsgFullSync:
		b.receberSincronizacaoCompleta(msg)
		conn.Close()
	case models.MsgReqDrone:
		if msg.SenderID == 0 {
			var droneID string
			if s, ok := msg.Payload.(string); ok {
				droneID = s
			} else {
				pBytes, _ := json.Marshal(msg.Payload)
				json.Unmarshal(pBytes, &droneID)
			}
			b.solicitarMissao(conn, droneID) // Endpoint para Drones
		} else {
			b.responderRicartAgrawala(msg) // Endpoint para P2P (Brokers)
			conn.Close()
		}
	case models.MsgReplyOK:
		b.processarReplyOK(msg)
		conn.Close()
	case models.MsgDroneHeartbeat:
		b.processarHeartbeat(msg)
		conn.Close()
	case models.MsgDroneConcluido:
		b.processarConclusaoDrone(msg)
		conn.Close()
	case models.MsgConsultaFila:
		b.responderConsultaFila(conn)
	case models.MsgNovoBloco:
		b.processarNovoBloco(msg)
		conn.Close()
	case models.MsgConsultaSaldo:
		var pubKey string
		pBytes, _ := json.Marshal(msg.Payload)
		json.Unmarshal(pBytes, &pubKey)
		b.responderConsultaSaldo(conn, pubKey)
	case models.MsgConsultaLedger:
		b.responderConsultaLedger(conn)
	default:
		conn.Close()
	}
}

// responderConsultaFila expõe um snapshot thread-safe da fila atual para clientes externos.
func (b *Broker) responderConsultaFila(conn net.Conn) {
	defer conn.Close()
	b.mu.Lock()
	copia := make([]models.Requisicao, len(b.ListaDistribuida))
	copy(copia, b.ListaDistribuida)
	b.mu.Unlock()

	resposta := models.RespostaFila{Requisicoes: copia}
	json.NewEncoder(conn).Encode(resposta)
}

// =========================================================
// Heartbeat e Recuperação de Falhas
// =========================================================

// processarHeartbeat atualiza a TTL de uma missão em curso.
// Implementa retransmissão de fofoca (Gossip) para evitar Split-Brain em Heartbeats perdidos.
func (b *Broker) processarHeartbeat(msg models.MensagemDistribuida) {
	pBytes, _ := json.Marshal(msg.Payload)
	var status models.DroneStatus
	json.Unmarshal(pBytes, &status)

	b.mu.Lock()
	for i := range b.ListaDistribuida {
		if b.ListaDistribuida[i].ID == status.MissionID {
			b.ListaDistribuida[i].UltimoHeartbeat = time.Now()
			break
		}
	}
	b.mu.Unlock()

	// Flagging: Assina a retransmissão do pacote para a malha se a origem foi o drone.
	if msg.SenderID == 0 {
		msg.SenderID = b.ID
		go b.broadcast(msg)
	}
}

// processarConclusaoDrone regista o fim da operação, atualiza o status e,
// SE o broker for o dono da missão, sela o Bloco na Blockchain.
func (b *Broker) processarConclusaoDrone(msg models.MensagemDistribuida) {
	pBytes, _ := json.Marshal(msg.Payload)
	var status models.DroneStatus
	json.Unmarshal(pBytes, &status)

	b.mu.Lock()
	var tarefaConcluida *models.Requisicao

	// 1. Encontra a missão e atualiza o estado local para parar o Watchdog
	for i := range b.ListaDistribuida {
		if b.ListaDistribuida[i].ID == status.MissionID {
			b.ListaDistribuida[i].Status = models.StatusConcluido
			tarefaConcluida = &b.ListaDistribuida[i]
			break
		}
	}

	if tarefaConcluida == nil {
		b.mu.Unlock()
		return
	}

	// =========================================================
	// REGRA DE AUTORIA PARA MINERAÇÃO (PREVENÇÃO DE FORK)
	// =========================================================
	// Todos os brokers ouvem o drone relatar a conclusão, mas APENAS
	// o broker "Sponsor" (dono da missão) tem autorização do consórcio
	// para escrever o laudo, selar o bloco e propagar a verdade absoluta.
	if tarefaConcluida.BrokerID != b.ID {
		b.mu.Unlock()
		return // Sou apenas um nó observador. Deixo a missão como concluída e espero o MsgNovoBloco.
	}

	// 2. Geração do Laudo de Auditoria Inalterável (Apenas para o Dono)
	laudo := models.LaudoAuditoria{
		DroneID:   tarefaConcluida.DroneID,
		Relatorio: "Patrulha concluída com sucesso. Nenhuma anomalia detetada.",
	}

	// 3. Determinar o Hash do Bloco Anterior (Encadeamento)
	hashAnt := "0000000000000000000000000000000000000000000000000000000000000000" // Genesis
	tamanhoCadeia := len(b.Blockchain.Cadeia)
	if tamanhoCadeia > 0 {
		hashAnt = b.Blockchain.Cadeia[tamanhoCadeia-1].HashAtual
	}

	// 4. Construção do Bloco
	novoBloco := models.Bloco{
		Index:        tamanhoCadeia + 1,
		Timestamp:    time.Now().UnixMilli(),
		MissaoID:     tarefaConcluida.ID,
		Transacao:    tarefaConcluida.Transacao,
		Laudo:        laudo,
		HashAnterior: hashAnt,
	}

	// 5. Prova de Integridade (Selagem criptográfica)
	novoBloco.HashAtual = novoBloco.CalcularHash()

	// 6. Adiciona ao Ledger local
	b.Blockchain.Cadeia = append(b.Blockchain.Cadeia, novoBloco)

	b.salvarLedgerLocal()

	fmt.Printf("[BLOCKCHAIN] Bloco #%d selado localmente! Hash: %s...\n", novoBloco.Index, novoBloco.HashAtual[:16])
	b.mu.Unlock()

	// 7. Propaga o novo Bloco para o Consórcio auditar
	go b.broadcast(models.MensagemDistribuida{
		Tipo:      models.MsgNovoBloco,
		SenderID:  b.ID,
		Timestamp: b.Relogio,
		Payload:   novoBloco,
	})
}

// processoWatchdogDrones é uma rotina assíncrona que varre missões em atendimento.
// Se uma missão violar a TTL (15s sem heartbeat), a tarefa sofre roll-back para status Pendente.
func (b *Broker) processoWatchdogDrones() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			var parabroadcast []models.MensagemDistribuida
			b.mu.Lock()
			for i := range b.ListaDistribuida {
				req := &b.ListaDistribuida[i]
				if req.Status != models.StatusEmAtendimento {
					continue
				}
				semHeartbeat := req.UltimoHeartbeat.IsZero() ||
					time.Since(req.UltimoHeartbeat) > 15*time.Second
				droneJaAssumiu := !req.IniciadoEm.IsZero() &&
					time.Since(req.IniciadoEm) > 15*time.Second

				if semHeartbeat && droneJaAssumiu {
					fmt.Printf("[WATCHDOG] Drone %-6s falhou | %s devolvida à fila | Alerta: \"%s\"\n",
						req.DroneID, req.ID, req.Descricao)
					req.Status = models.StatusPendente
					req.DroneID = ""
					req.IniciadoEm = time.Time{}
					req.UltimoHeartbeat = time.Time{}
					parabroadcast = append(parabroadcast, models.MensagemDistribuida{
						Tipo:     models.MsgSyncUpdate,
						SenderID: b.ID,
						Payload:  *req,
					})
				}
			}
			b.mu.Unlock()

			for _, bMsg := range parabroadcast {
				go b.broadcast(bMsg)
			}
		}
	}()
}

// =========================================================
// Handshakes e Sincronização Topológica
// =========================================================

// solicitarEntrada dispara o Handshake para o Seed. Inicia a entrada na topologia de rede.
func (b *Broker) solicitarEntrada(seedAddr string) {
	fmt.Printf("[JOIN] Tentando entrar na malha via semente: %s\n", seedAddr)
	conn, err := net.DialTimeout("tcp", seedAddr, 5*time.Second)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao conectar na semente %s: %v\n", seedAddr, err)
		return
	}
	defer conn.Close()

	b.mu.Lock()
	b.Relogio++
	ts := b.Relogio
	b.mu.Unlock()

	joinMsg := models.MensagemDistribuida{
		Tipo:      models.MsgJoin,
		SenderID:  b.ID,
		Timestamp: ts,
		Payload: models.JoinRequest{
			ID:   b.ID,
			Addr: b.MeuEndereco,
		},
	}

	if err := json.NewEncoder(conn).Encode(joinMsg); err != nil {
		return
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var ack models.MensagemDistribuida
	if err := json.NewDecoder(conn).Decode(&ack); err != nil {
		fmt.Printf("[ERRO] Sem resposta síncrona do seed %s: %v\n", seedAddr, err)
		return
	}

	if ack.Tipo == models.MsgJoinACK {
		var seedInfo models.JoinRequest
		pBytes, _ := json.Marshal(ack.Payload)
		json.Unmarshal(pBytes, &seedInfo)

		b.mu.Lock()
		b.Peers[seedInfo.ID] = seedInfo.Addr
		b.FalhasConsecutivas[seedInfo.ID] = 0
		fmt.Printf("[JOIN] Seed %d (%s) adicionado aos Peers\n", seedInfo.ID, seedInfo.Addr)
		b.mu.Unlock()
	}
}

// processarJoin cadastra novos nós, dispara retransmissão de IP para a malha
// e devolve um ACK síncrono ou assíncrono dependendo da rota de entrada do pacote.
func (b *Broker) processarJoin(msg models.MensagemDistribuida, conn net.Conn) {
	pBytes, _ := json.Marshal(msg.Payload)
	var reqJoin models.JoinRequest
	json.Unmarshal(pBytes, &reqJoin)

	if reqJoin.ID == b.ID {
		conn.Close()
		return
	}

	b.mu.Lock()
	_, jaConhece := b.Peers[reqJoin.ID]
	if !jaConhece {
		b.Peers[reqJoin.ID] = reqJoin.Addr
		b.FalhasConsecutivas[reqJoin.ID] = 0
		fmt.Printf("[P2P] Novo Broker %d adicionado: %s\n", reqJoin.ID, reqJoin.Addr)
	}
	b.mu.Unlock()

	ack := models.MensagemDistribuida{
		Tipo:      models.MsgJoinACK,
		SenderID:  b.ID,
		Timestamp: b.Relogio,
		Payload:   models.JoinRequest{ID: b.ID, Addr: b.MeuEndereco},
	}

	// Handshake Síncrono direto no Socket aberto, ou Assíncrono via Discovery (Broadcast Gossip)
	if msg.SenderID == reqJoin.ID {
		json.NewEncoder(conn).Encode(ack)
	} else {
		go func(targetAddr string) {
			connDir, err := net.DialTimeout("tcp", targetAddr, 2*time.Second)
			if err == nil {
				json.NewEncoder(connDir).Encode(ack)
				connDir.Close()
			}
		}(reqJoin.Addr)
	}

	conn.Close()

	if !jaConhece {
		msgPropagada := msg
		msgPropagada.SenderID = b.ID
		go b.broadcast(msgPropagada)

		if msg.SenderID == reqJoin.ID {
			go b.enviarEstadoParaNovato(reqJoin.Addr)
		}
	}
}

// processarJoinACK finaliza a rotina bidirecional de Handshake anexando a semente à lista local.
func (b *Broker) processarJoinACK(msg models.MensagemDistribuida) {
	pBytes, _ := json.Marshal(msg.Payload)
	var info models.JoinRequest
	json.Unmarshal(pBytes, &info)

	b.mu.Lock()
	if _, existe := b.Peers[info.ID]; !existe {
		b.Peers[info.ID] = info.Addr
		b.FalhasConsecutivas[info.ID] = 0
		fmt.Printf("[P2P] Peer %d (%s) adicionado via ACK\n", info.ID, info.Addr)
	}
	b.mu.Unlock()
}

// enviarEstadoParaNovato provê a Transferência de Estado (FullSync) ao nó recém-aceito.
func (b *Broker) enviarEstadoParaNovato(addr string) {
	time.Sleep(500 * time.Millisecond)
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	b.mu.Lock()
	msg := models.MensagemDistribuida{
		Tipo:      models.MsgFullSync,
		SenderID:  b.ID,
		Timestamp: b.Relogio,
		Payload:   b.ListaDistribuida,
	}
	b.mu.Unlock()

	json.NewEncoder(conn).Encode(msg)
}

// receberSincronizacaoCompleta processa o snapshot e substitui a malha local.
func (b *Broker) receberSincronizacaoCompleta(msg models.MensagemDistribuida) {
	b.mu.Lock()
	defer b.mu.Unlock()

	pBytes, _ := json.Marshal(msg.Payload)
	var listaNova []models.Requisicao
	json.Unmarshal(pBytes, &listaNova)

	b.ListaDistribuida = listaNova
	b.ordenarLista()
	fmt.Printf("[SYNC] Lista sincronizada: %d missões\n", len(listaNova))
}

// =========================================================
// Processamento de Objetos de Alerta
// =========================================================

// processarNovoAlerta ingere uma requisição gerada pela rede de Sensores/Clientes.
func (b *Broker) processarNovoAlerta(msg models.MensagemDistribuida) {
	var req models.Requisicao
	pBytes, _ := json.Marshal(msg.Payload)
	json.Unmarshal(pBytes, &req)

	// =========================================================
	// 1. AUDITORIA DE SEGURANÇA (CRIPTOGRAFIA)
	// =========================================================
	if err := b.ValidarAssinatura(req.Transacao); err != nil {
		fmt.Printf("[ALERTA DE SEGURANÇA] Pacote rejeitado! Motivo: %v\n", err)
		return // Descarta o pacote fraudulento
	}

	// =========================================================
	// 2. AUDITORIA FINANCEIRA (PREVENÇÃO DE DUPLO GASTO)
	// =========================================================
	saldoDisponivel := b.CalcularSaldoEfetivo(req.Transacao.CompanhiaPubKey)
	if saldoDisponivel < req.Transacao.Valor {
		fmt.Printf("[FRAUDE DETECTADA] Companhia tentou gastar %.2f, mas possui %.2f créditos!\n",
			req.Transacao.Valor, saldoDisponivel)
		return // Descarta pacote sem saldo
	}

	// Se passou nas duas auditorias, aceita na Mempool
	b.mu.Lock()
	origemLocal := msg.Timestamp == 0
	if origemLocal {
		b.Relogio++
		req.Timestamp = b.Relogio
		req.BrokerID = b.ID
	} else {
		req.Timestamp = msg.Timestamp
		req.BrokerID = msg.SenderID
	}
	req.Status = models.StatusPendente
	b.ListaDistribuida = append(b.ListaDistribuida, req)
	b.ordenarLista()
	b.mu.Unlock()

	if origemLocal {
		msg.Timestamp = req.Timestamp
		msg.SenderID = b.ID
		go b.broadcast(msg)
	}
}

// processarUpdateStatus processa flags atômicas de atualização do fluxo da missão.
func (b *Broker) processarUpdateStatus(msg models.MensagemDistribuida) {
	var update models.Requisicao
	pBytes, _ := json.Marshal(msg.Payload)
	json.Unmarshal(pBytes, &update)

	b.mu.Lock()
	defer b.mu.Unlock()
	for i, r := range b.ListaDistribuida {
		if r.ID == update.ID {
			b.ListaDistribuida[i].Status = update.Status
			b.ListaDistribuida[i].DroneID = update.DroneID

			// Protocolo Fail-Fast: Cancela requisição de alocação P2P local
			// se missão for assumida preempetivamente por outro nó.
			if update.Status == models.StatusEmAtendimento {
				b.ListaDistribuida[i].IniciadoEm = time.Now()
				b.ListaDistribuida[i].UltimoHeartbeat = time.Now()

				if pend, existe := pendencias[update.ID]; existe {
					fmt.Printf("[R-A] A missão %s foi assumida por outro. Recalculando...\n", update.ID)
					droneConn := pend.DroneConn
					droneID := pend.DroneID
					delete(pendencias, update.ID)
					go b.solicitarMissao(droneConn, droneID)
				}
			}
			break
		}
	}
}

// =========================================================
// Consenso Distribuído (Ricart-Agrawala Modificado)
// =========================================================

// solicitarMissao inicia a varredura local por pendências elegíveis.
// Se encontradas, dispara requisições de OK (multicast) exigindo consenso da malha.
func (b *Broker) solicitarMissao(conn net.Conn, droneID string) {
	b.mu.Lock()
	var alvo *models.Requisicao

	// NOVA REGRA DE AUTORIA:
	// O Broker varre a lista distribuída, mas APENAS tenta capturar um drone
	// para as missões que ele mesmo originou (BrokerID == b.ID).
	for i := range b.ListaDistribuida {
		req := &b.ListaDistribuida[i]

		if req.Status == models.StatusPendente && req.BrokerID == b.ID {
			// Prevenção de Concorrência Reentrante
			if _, jaNegociando := pendencias[req.ID]; !jaNegociando {
				alvo = req
				break
			}
		}
	}

	if alvo == nil {
		b.mu.Unlock()
		// Responde ao drone com uma requisição vazia (liberando-o para tentar outro broker)
		json.NewEncoder(conn).Encode(models.Requisicao{})
		conn.Close()
		return
	}

	b.Relogio++
	timestampLocal := b.Relogio
	idMissao := alvo.ID

	totalPeers := 0
	for id := range b.Peers {
		if id != b.ID {
			totalPeers++
		}
	}

	pendencia := &PendenciaAgrawala{
		MissionID:      idMissao,
		TimestampLocal: timestampLocal,
		Respostas:      make(map[int]bool),
		TotalEsperado:  totalPeers,
		DroneConn:      conn,
		DroneID:        droneID,
	}
	pendencias[idMissao] = pendencia

	if totalPeers == 0 {
		fmt.Printf("[R-A] Sem peers. Missão %s autorizada de imediato.\n", idMissao)
		go b.confirmarMissaoAoDrone(pendencia)
		delete(pendencias, idMissao)
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()

	// Watchdog Concorrente para timeout da requisição P2P
	go func(idM string) {
		time.Sleep(5 * time.Second)
		b.mu.Lock()
		if pend, existe := pendencias[idM]; existe {
			fmt.Printf("[WATCHDOG] Timeout com %d/%d respostas. Confirmando missão %s.\n",
				len(pend.Respostas), pend.TotalEsperado, idM)
			go b.confirmarMissaoAoDrone(pend)
			delete(pendencias, idM)
		}
		b.mu.Unlock()
	}(idMissao)

	reqMsg := models.MensagemDistribuida{
		Tipo:      models.MsgReqDrone,
		SenderID:  b.ID,
		Timestamp: timestampLocal,
		Payload:   idMissao,
	}

	entregues, _ := b.broadcast(reqMsg)
	falhasDeEnvio := totalPeers - entregues

	b.mu.Lock()
	if pend, existe := pendencias[idMissao]; existe {
		pend.TotalEsperado -= falhasDeEnvio

		if pend.TotalEsperado <= 0 {
			fmt.Printf("[R-A] Todos os peers inativos. Missão %s autorizada.\n", idMissao)
			go b.confirmarMissaoAoDrone(pend)
			delete(pendencias, idMissao)
		} else if len(pend.Respostas) >= pend.TotalEsperado {
			fmt.Printf("[R-A] Todos os OKs recebidos (%d/%d). Confirmando missão %s.\n", len(pend.Respostas), pend.TotalEsperado, idMissao)
			go b.confirmarMissaoAoDrone(pend)
			delete(pendencias, idMissao)
		}
	}
	b.mu.Unlock()
}

// responderRicartAgrawala atua como receptor de permissões. Implementa a lógica
// de desempate (Lamport) em caso de Lock Distribuído sobre o mesmo recurso.
func (b *Broker) responderRicartAgrawala(msg models.MensagemDistribuida) {
	missionID, ok := msg.Payload.(string)
	if !ok {
		return
	}

	b.mu.Lock()
	okParaEnviar := true

	if pendLocal, existe := pendencias[missionID]; existe {
		meTemPrioridade := pendLocal.TimestampLocal < msg.Timestamp ||
			(pendLocal.TimestampLocal == msg.Timestamp && b.ID < msg.SenderID)
		if meTemPrioridade {
			okParaEnviar = false
			fmt.Printf("[R-A] Negando OK para Broker %d (nossa TS=%d, remota=%d)\n",
				msg.SenderID, pendLocal.TimestampLocal, msg.Timestamp)
		}
	}

	destAddr := ""
	if addr, existe := b.Peers[msg.SenderID]; existe {
		destAddr = addr
	}
	ts := b.Relogio
	b.mu.Unlock()

	if okParaEnviar {
		if destAddr != "" {
			b.enviarOKParaAddr(destAddr, ts, missionID)
		} else {
			fmt.Printf("[R-A] ALERTA: Não consigo enviar OK para Broker %d porque o IP dele sumiu da minha lista!\n", msg.SenderID)
		}
	}
}

// enviarOKParaAddr encerra as transações TCP relativas a Replies (votos de aprovação).
func (b *Broker) enviarOKParaAddr(addr string, ts int64, missionID string) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	json.NewEncoder(conn).Encode(models.MensagemDistribuida{
		Tipo:      models.MsgReplyOK,
		SenderID:  b.ID,
		Timestamp: ts,
		Payload:   missionID,
	})
}

// processarReplyOK é o sink collector para o quórum R-A. Se o array contido
// no dicionário bater com TotalEsperado, despacha a missão.
func (b *Broker) processarReplyOK(msg models.MensagemDistribuida) {
	b.mu.Lock()
	defer b.mu.Unlock()

	missionID, ok := msg.Payload.(string)
	if !ok {
		pBytes, _ := json.Marshal(msg.Payload)
		json.Unmarshal(pBytes, &missionID)
		if missionID == "" {
			return
		}
	}

	p, existe := pendencias[missionID]
	if !existe {
		return
	}

	p.Respostas[msg.SenderID] = true
	votosAtuais := len(p.Respostas)

	fmt.Printf("[R-A] OK de Broker %d para missão %s (%d/%d)\n",
		msg.SenderID, missionID, votosAtuais, p.TotalEsperado)

	if p.TotalEsperado > 0 && votosAtuais >= p.TotalEsperado {
		go b.confirmarMissaoAoDrone(p)
		delete(pendencias, missionID)
	}
}

// confirmarMissaoAoDrone comuta as estruturas e descarrega a tarefa ao Drone ativo
// fechando o File Descriptor (Conn).
func (b *Broker) confirmarMissaoAoDrone(p *PendenciaAgrawala) {
	defer p.DroneConn.Close()

	b.mu.Lock()
	var tarefa models.Requisicao
	var encontrou bool
	for i, r := range b.ListaDistribuida {
		if r.ID == p.MissionID {
			b.ListaDistribuida[i].Status = models.StatusEmAtendimento
			b.ListaDistribuida[i].DroneID = p.DroneID
			b.ListaDistribuida[i].IniciadoEm = time.Now()
			b.ListaDistribuida[i].UltimoHeartbeat = time.Now()
			tarefa = b.ListaDistribuida[i]
			encontrou = true
			break
		}
	}
	b.mu.Unlock()

	if !encontrou {
		return
	}

	if err := json.NewEncoder(p.DroneConn).Encode(tarefa); err != nil {
		fmt.Printf("[DRONE] Erro ao enviar missão ao drone %s: %v\n", p.DroneID, err)
		return
	}

	fmt.Printf("[DESPACHO] Drone %-6s <- %s | Alerta: \"%s\" | Setor %d | Prioridade %d\n",
		p.DroneID, p.MissionID, tarefa.Descricao, tarefa.Setor, tarefa.Prioridade)

	go b.broadcast(models.MensagemDistribuida{
		Tipo:      models.MsgSyncUpdate,
		SenderID:  b.ID,
		Timestamp: b.Relogio,
		Payload:   tarefa,
	})
}

// =========================================================
// Envelhecimento (Aging)
// =========================================================

// processoEnvelhecimento rotaciona uma Ticker Function de elevação automática
// para prevenir Starvation de requisições de baixa prioridade em fila de espera.
func (b *Broker) processoEnvelhecimento() {
	ticker := time.NewTicker(20 * time.Second)
	go func() {
		for range ticker.C {
			b.mu.Lock()
			houveMudanca := false
			for i := range b.ListaDistribuida {
				req := &b.ListaDistribuida[i]
				if req.Status == models.StatusPendente && req.Prioridade < 5 {
					if time.Since(req.CreatedAt) > 45*time.Second {
						req.Prioridade++
						houveMudanca = true
						fmt.Printf("[AGING] Requisição %s → Prioridade %d\n", req.ID, req.Prioridade)
					}
				}
			}
			if houveMudanca {
				b.ordenarLista()
			}
			b.mu.Unlock()
		}
	}()
}

// =========================================================
// Topologia e Reestabelecimento (Heartbeat Loop)
// =========================================================

// processoReconexao é a malha infinita (Polling) para auto-cura do cluster.
// Realiza chamadas agressivas iterando o array originário do .env se houver perda de malha.
func (b *Broker) processoReconexao(seeds []string) {
	time.Sleep(2 * time.Second)
	for {
		b.mu.Lock()
		temPeer := len(b.Peers) > 0
		b.mu.Unlock()

		if !temPeer {
			fmt.Printf("[RECONEXAO] Sem peers. Tentando seeds: %v\n", seeds)
			for _, seed := range seeds {
				b.solicitarEntrada(seed)
			}

			b.mu.Lock()
			temPeer = len(b.Peers) > 0
			b.mu.Unlock()

			if temPeer {
				fmt.Printf("[RECONEXAO] Conectado com sucesso à malha!\n")
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// processoAdocaoOrfaos varre a lista distribuída buscando missões de brokers que caíram.
// Para evitar concorrência no roubo de missões, apenas o broker sobrevivente com o MENOR ID
// assume a autoria da missão órfã.
func (b *Broker) processoAdocaoOrfaos() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			var adotadas []models.MensagemDistribuida

			b.mu.Lock()
			for i := range b.ListaDistribuida {
				req := &b.ListaDistribuida[i]

				// Se a missão é de OUTRO broker e está pendente...
				if req.Status == models.StatusPendente && req.BrokerID != b.ID {
					_, donoVivo := b.Peers[req.BrokerID]

					// ...e o dono dela não está mais ativo na nossa lista de Peers
					if !donoVivo {
						// LÓGICA DE DESEMPATE: Quem herda a missão?
						// Resposta: O nó sobrevivente com o menor ID.
						menorIdSobrevivente := b.ID
						for peerID := range b.Peers {
							if peerID < menorIdSobrevivente {
								menorIdSobrevivente = peerID
							}
						}

						// Se eu sou o nó sobrevivente com o menor ID, eu adoto a missão.
						if b.ID == menorIdSobrevivente {
							fmt.Printf("[FAILOVER] Broker %d caiu! Adotando missão %s...\n", req.BrokerID, req.ID)
							req.BrokerID = b.ID // Eu assumo a autoria

							adotadas = append(adotadas, models.MensagemDistribuida{
								Tipo:      models.MsgSyncUpdate,
								SenderID:  b.ID,
								Timestamp: b.Relogio,
								Payload:   *req,
							})
						}
					}
				}
			}
			b.mu.Unlock()

			// Faz o broadcast das missões que foram sequestradas para a malha atualizar
			for _, msg := range adotadas {
				go b.broadcast(msg)
			}
		}
	}()
}

// CalcularSaldoEfetivo varre o histórico imutável (Blockchain) e as missões
// ativas (ListaDistribuida) para determinar o saldo real de uma companhia,
// blindando o sistema contra o ataque de Duplo Gasto e auditando subsídios.
func (b *Broker) CalcularSaldoEfetivo(companhiaPubKey string) float64 {
	// A cota inicial que cada nação/companhia recebe do Consórcio.
	saldo := 200.0

	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. DEDUÇÃO E ADIÇÃO DA BLOCKCHAIN (Gastos e Subsídios Confirmados)
	// Varre todos os blocos selados do livro-razão.
	for _, bloco := range b.Blockchain.Cadeia {
		// A. Verifica se é uma transação da própria companhia (Gasto)
		if bloco.Transacao.CompanhiaPubKey == companhiaPubKey {
			saldo -= bloco.Transacao.Valor
		}

		// B. Verifica se é uma transação de Subsídio Global (do Consórcio)
		// Aqui adicionamos o subsídio ao saldo da companhia.
		if bloco.Transacao.CompanhiaPubKey == "SUBSIDIO_CONSORCIO_ORMUZ" {
			// Como o valor do subsídio foi registrado como negativo (-50.0),
			// subtrair um negativo resulta em uma soma (saldo + 50.0).
			saldo -= bloco.Transacao.Valor
		}
	}

	// 2. DEDUÇÃO DA MEMPOOL (Fundos Congelados)
	// Varre a fila de missões em andamento. Se a empresa pediu um drone que
	// ainda não voltou, o dinheiro daquela escolta fica bloqueado (cativo).
	for _, req := range b.ListaDistribuida {
		if req.Transacao.CompanhiaPubKey == companhiaPubKey && req.Status != models.StatusConcluido {
			saldo -= req.Transacao.Valor
		}
	}

	return saldo
}

// ValidarAssinatura reconstrói a chave pública e verifica o selo ECDSA
func (b *Broker) ValidarAssinatura(tx models.TransacaoToken) error {
	// 1. Descodificar a Chave Pública de Hex para Bytes
	pubKeyBytes, err := hex.DecodeString(tx.CompanhiaPubKey)
	if err != nil {
		return errors.New("formato de chave pública inválido")
	}

	// 2. Reconstruir a estrutura ecdsa.PublicKey
	x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
	if x == nil {
		return errors.New("chave pública corrompida ou inválida")
	}
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// 3. Descodificar a Assinatura de Hex para Bytes
	assinaturaBytes, err := hex.DecodeString(tx.Assinatura)
	if err != nil {
		return errors.New("formato de assinatura inválido")
	}

	// 4. Regenerar o Hash original dos dados
	hash := tx.GerarHashDados()

	// 5. O Teste Criptográfico Final
	valido := ecdsa.VerifyASN1(pubKey, hash, assinaturaBytes)
	if !valido {
		return errors.New("VERIFICAÇÃO FALHOU: Transação forjada ou adulterada em trânsito")
	}

	return nil
}

// processoGeradorCreditos injeta créditos periodicamente na rede.
// Utiliza Eleição Dinâmica de Líder (Menor ID ativo) para evitar Forks na Blockchain.
func (b *Broker) processoGeradorCreditos() {
	ticker := time.NewTicker(120 * time.Second)
	go func() {
		for range ticker.C {
			b.mu.Lock()

			// =========================================================
			// 1. ELEIÇÃO DINÂMICA DO LÍDER (BANCO CENTRAL)
			// =========================================================
			// Verifica quem é o nó ativo com o menor ID na malha P2P atual.
			liderAtual := b.ID
			for peerID := range b.Peers {
				if peerID < liderAtual {
					liderAtual = peerID
				}
			}

			// Se este Broker NÃO for o líder eleito, ele aborta a geração do bloco
			// e fica apenas aguardando receber o bloco já minerado pelo líder.
			if b.ID != liderAtual {
				b.mu.Unlock()
				continue
			}

			// Se o código chegou aqui, este Broker é o Líder!
			fmt.Printf("[ECONOMIA] Sou o Líder atual (Broker %d)! Minerando novo subsídio operacional...\n", b.ID)

			// =========================================================
			// 2. Criar Transação do Sistema (O "Consórcio" dá 50 créditos)
			// =========================================================
			// Nota: Em produção, você usaria uma chave privada dedicada ao "Banco Central"
			tx := models.TransacaoToken{
				CompanhiaPubKey: "SUBSIDIO_CONSORCIO_ORMUZ",
				Valor:           -50.0, // Valor negativo para representar entrada de crédito
				Timestamp:       time.Now().UnixMilli(),
				Assinatura:      "SISTEMA_VALIDADO",
			}

			// =========================================================
			// 3. Construir o Bloco de Subsídio
			// =========================================================
			tamanhoCadeia := len(b.Blockchain.Cadeia)
			hashAnt := "0000000000000000000000000000000000000000000000000000000000000000"
			if tamanhoCadeia > 0 {
				hashAnt = b.Blockchain.Cadeia[tamanhoCadeia-1].HashAtual
			}

			novoBloco := models.Bloco{
				Index:        tamanhoCadeia + 1,
				Timestamp:    time.Now().UnixMilli(),
				MissaoID:     fmt.Sprintf("SUBSIDIO-%d", time.Now().Unix()),
				Transacao:    tx,
				Laudo:        models.LaudoAuditoria{DroneID: "SYSTEM", Relatorio: "Injeção de subsídio."},
				HashAnterior: hashAnt,
			}
			novoBloco.HashAtual = novoBloco.CalcularHash()

			// =========================================================
			// 4. Adicionar ao Ledger Local e salvar no disco
			// =========================================================
			b.Blockchain.Cadeia = append(b.Blockchain.Cadeia, novoBloco)
			b.salvarLedgerLocal()
			b.mu.Unlock()

			// =========================================================
			// 5. Propagar (Broadcast) o novo bloco para a malha validar
			// =========================================================
			go b.broadcast(models.MensagemDistribuida{
				Tipo:      models.MsgNovoBloco,
				SenderID:  b.ID,
				Timestamp: time.Now().UnixMilli(),
				Payload:   novoBloco,
			})
		}
	}()
}

// processarNovoBloco recebe um bloco da malha P2P, audita a sua integridade e anexa ao Ledger.
func (b *Broker) processarNovoBloco(msg models.MensagemDistribuida) {
	var blocoRecebido models.Bloco
	pBytes, _ := json.Marshal(msg.Payload)
	json.Unmarshal(pBytes, &blocoRecebido)

	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. Auditoria de Integridade (O bloco foi adulterado em trânsito?)
	hashCalculado := blocoRecebido.CalcularHash()
	if hashCalculado != blocoRecebido.HashAtual {
		fmt.Printf("[SEGURANÇA] Bloco rejeitado! Hash inválido. Tentativa de adulteração detetada.\n")
		return
	}

	// 2. Auditoria de Encadeamento (O bloco encaixa na nossa cadeia?)
	hashAntEsperado := "0000000000000000000000000000000000000000000000000000000000000000"
	tamanhoCadeia := len(b.Blockchain.Cadeia)
	if tamanhoCadeia > 0 {
		hashAntEsperado = b.Blockchain.Cadeia[tamanhoCadeia-1].HashAtual
	}

	if blocoRecebido.HashAnterior != hashAntEsperado {
		fmt.Printf("[SEGURANÇA] Bloco rejeitado! Desincronização (Falta de consenso no Hash Anterior).\n")
		return
	}

	// 3. Tudo válido. Anexamos ao Ledger Imutável.
	b.Blockchain.Cadeia = append(b.Blockchain.Cadeia, blocoRecebido)

	b.salvarLedgerLocal()

	// 4. Limpeza da Mempool: Removemos a missão da fila ativa, pois já está no histórico
	for i := range b.ListaDistribuida {
		if b.ListaDistribuida[i].ID == blocoRecebido.MissaoID {
			b.ListaDistribuida[i].Status = models.StatusConcluido
			break
		}
	}

	fmt.Printf("[LEDGER] Bloco #%d anexado com sucesso via P2P. Missão %s imutável.\n",
		blocoRecebido.Index, blocoRecebido.MissaoID)
}

// getNomeArquivoLedger retorna o nome do ficheiro único para este broker
func (b *Broker) getNomeArquivoLedger() string {
	return fmt.Sprintf("data/ledger_broker_%d.json", b.ID)
}

// carregarLedgerLocal lê o ficheiro do disco (se existir) e carrega a Blockchain para a memória
func (b *Broker) carregarLedgerLocal() {
	nomeFicheiro := b.getNomeArquivoLedger()

	// Verifica se o ficheiro existe
	if _, err := os.Stat(nomeFicheiro); os.IsNotExist(err) {
		fmt.Println("[PERSISTÊNCIA] Nenhum ledger local encontrado. Iniciando blockchain vazia.")
		b.Blockchain.Cadeia = []models.Bloco{}
		return
	}

	dados, err := ioutil.ReadFile(nomeFicheiro)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao ler ledger local: %v\n", err)
		return
	}

	if err := json.Unmarshal(dados, &b.Blockchain.Cadeia); err != nil {
		fmt.Printf("[ERRO] Falha ao descodificar ledger local: %v\n", err)
		return
	}

	fmt.Printf("[PERSISTÊNCIA] Ledger carregado com sucesso! %d blocos encontrados.\n", len(b.Blockchain.Cadeia))
}

// salvarLedgerLocal reescreve o ficheiro JSON com a cadeia atualizada
func (b *Broker) salvarLedgerLocal() {
	nomeFicheiro := b.getNomeArquivoLedger()

	// Formatação "Indent" para ficar bonito e legível para a apresentação
	dados, err := json.MarshalIndent(b.Blockchain.Cadeia, "", "  ")
	if err != nil {
		fmt.Printf("[ERRO] Falha ao serializar blockchain: %v\n", err)
		return
	}

	if err := ioutil.WriteFile(nomeFicheiro, dados, 0644); err != nil {
		fmt.Printf("[ERRO] Falha ao salvar ficheiro ledger: %v\n", err)
	}
}

// responderConsultaSaldo expõe o resultado financeiro auditado para o cliente
func (b *Broker) responderConsultaSaldo(conn net.Conn, pubKey string) {
	defer conn.Close()
	// Reaproveitamos a função blindada que varre a Blockchain
	saldo := b.CalcularSaldoEfetivo(pubKey)
	json.NewEncoder(conn).Encode(saldo)
}

// responderConsultaLedger envia a cópia fiel da Blockchain (persistida no disco) para o cliente
func (b *Broker) responderConsultaLedger(conn net.Conn) {
	defer conn.Close()
	b.mu.Lock()
	// Cria uma cópia thread-safe da cadeia de blocos
	copia := make([]models.Bloco, len(b.Blockchain.Cadeia))
	copy(copia, b.Blockchain.Cadeia)
	b.mu.Unlock()

	json.NewEncoder(conn).Encode(copia)
}

// buscarAtualizacaoDeLedger conecta-se aos nós sementes/vizinhos ao iniciar
// e verifica se eles possuem um histórico de blocos maior (mais atualizado).
// Se sim, ele faz o download e sobrescreve o seu próprio disco para "alcançar" a rede.
func (b *Broker) buscarAtualizacaoDeLedger(seeds []string) {
	fmt.Println("[SYNC] Aguardando estabilização da rede P2P (Warm-up)...")

	// 1. ESPERA INTELIGENTE: Tenta achar vizinhos por até 10 segundos
	for i := 0; i < 10; i++ {
		b.mu.Lock()
		temPeer := len(b.Peers) > 0
		b.mu.Unlock()

		if temPeer {
			break // Achou alguém! Pode sair da espera.
		}
		time.Sleep(1 * time.Second)
	}

	// 2. Dá mais 2 segundinhos de "gordura" para que todos os Handshakes e ACKs terminem
	time.Sleep(2 * time.Second)

	b.mu.Lock()
	tamanhoLocal := len(b.Blockchain.Cadeia)
	b.mu.Unlock()

	fmt.Printf("[SYNC] Iniciando protocolo de recuperação. Tamanho local: %d blocos...\n", tamanhoLocal)

	maiorCadeia := []models.Bloco{}
	maiorTamanho := tamanhoLocal

	// Interroga todos os nós vizinhos (usando a lista de seeds que já conhecemos)
	for _, peerAddr := range seeds {
		peerAddr = strings.TrimSpace(peerAddr)

		// CORREÇÃO AQUI: b.MeuEndereco em vez de b.Endereco
		if peerAddr == "" || peerAddr == b.MeuEndereco {
			continue // Não pergunta a si mesmo
		}

		conn, err := net.DialTimeout("tcp", peerAddr, 2*time.Second)
		if err != nil {
			continue // Nó vizinho offline, ignora
		}

		// Pede a Blockchain completa para o vizinho
		msg := models.MensagemDistribuida{
			Tipo:      models.MsgConsultaLedger,
			SenderID:  b.ID,
			Timestamp: time.Now().UnixMilli(),
		}
		json.NewEncoder(conn).Encode(msg)

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		var cadeiaRecebida []models.Bloco
		err = json.NewDecoder(conn).Decode(&cadeiaRecebida)
		conn.Close()

		// Se recebeu com sucesso e é maior que a local, guarda como a maior encontrada
		if err == nil && len(cadeiaRecebida) > maiorTamanho {
			maiorTamanho = len(cadeiaRecebida)
			maiorCadeia = cadeiaRecebida
		}
	}

	// Se encontrou uma cadeia maior na rede, atualiza o seu próprio cérebro e disco
	if len(maiorCadeia) > tamanhoLocal {
		b.mu.Lock()
		b.Blockchain.Cadeia = maiorCadeia
		b.salvarLedgerLocal() // Grava a atualização por cima do ficheiro velho no Windows/Linux
		b.mu.Unlock()
		fmt.Printf("[SYNC-SUCESSO] Ledger desatualizado! Fiz o download de %d novos blocos da rede.\n", len(maiorCadeia)-tamanhoLocal)
	} else {
		fmt.Println("[SYNC] O Ledger local já está atualizado com a malha.")
	}
}

// =========================================================
// Ponto de Entrada (Entrypoint Node)
// =========================================================

// main constrói as entidades e realiza o mapeamento das v-envs em estruturas persistentes.
func main() {
	id, _ := strconv.Atoi(os.Getenv("BROKER_ID"))
	porta := os.Getenv("BROKER_PORT")
	peersStr := os.Getenv("PEERS")
	seedAddr := os.Getenv("SEED_ADDR")
	seedAddrs := os.Getenv("SEED_ADDRS")
	meuEndereco := os.Getenv("MY_ADDR")

	if porta == "" {
		porta = ":9000"
	}

	var seeds []string
	raw := seedAddrs
	if raw == "" {
		raw = seedAddr
	}
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" && s != meuEndereco {
			seeds = append(seeds, s)
		}
	}

	broker := &Broker{
		ID:                 id,
		Porta:              porta,
		MeuEndereco:        meuEndereco,
		Peers:              make(map[int]string),
		FalhasConsecutivas: make(map[int]int),
	}

	if peersStr != "" {
		for i, p := range strings.Split(peersStr, ",") {
			p = strings.TrimSpace(p)
			parts := strings.Split(p, "=")
			if len(parts) == 2 {
				pID, _ := strconv.Atoi(parts[0])
				broker.Peers[pID] = parts[1]
			} else if p != "" {
				broker.Peers[i+2] = p
			}
		}
	}

	// CARREGA A MEMÓRIA DO DISCO ANTES DE INICIAR
	broker.carregarLedgerLocal()

	// Inicialização atrasada para garantir sincronia física nos containers
	//if len(seeds) > 0 && meuEndereco != "" {
	//	go broker.processoReconexao(seeds)
	//	fmt.Println(">>> Aguardando estabilização da malha P2P (3s)...")
	//	time.Sleep(3 * time.Second)
	//}

	fmt.Printf(">>> Broker %d iniciando na porta %s | externo: %s\n", broker.ID, porta, meuEndereco)
	broker.Iniciar(seeds)
}
