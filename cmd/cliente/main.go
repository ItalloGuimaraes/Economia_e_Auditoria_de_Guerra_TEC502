package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	mathrand "math/rand"
	"net"
	"ormuz_distribuido/internal/models"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	brokerAddrs     []string
	chavePrivada    *ecdsa.PrivateKey
	companhiaPubKey string
	brokerGateway   string // O nó primário escolhido pela Companhia (O API Gateway)
	nomeCompanhia   string // O nome digitado pelo usuário (Login)
)

func main() {
	addrsRaw := os.Getenv("BROKER_ADDRS")
	if addrsRaw == "" {
		addrsRaw = "broker-1:9000,broker-2:9000,broker-3:9000,broker-4:9000"
	}

	for _, a := range strings.Split(addrsRaw, ",") {
		brokerAddrs = append(brokerAddrs, strings.TrimSpace(a))
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║         SISTEMA DE MONITORAMENTO — ESTREITO DE ORMUZ         ║")
	fmt.Println("║                   Terminal da Companhia                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	scanner := bufio.NewScanner(os.Stdin)

	// 1. TELA DE LOGIN: Coleta o nome da Companhia dinamicamente
	for {
		fmt.Print("\n  ▶ Digite o nome da sua Companhia de Navegação: ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			// Formata para usar como nome do arquivo (ex: "MSC Cruzeiros" vira "msc_cruzeiros")
			nomeCompanhia = strings.ToLower(strings.ReplaceAll(input, " ", "_"))
			break
		}
	}

	// 2. CARTEIRA: Carrega ou cria a Wallet persistente baseada no nome digitado
	inicializarIdentidadePersistente()

	// 3. CONEXÃO DE INFRAESTRUTURA: A Companhia liga-se exclusivamente a este nó.
	fmt.Println("\n  ▶ Selecione a sua infraestrutura de Gateway (Broker Oficial):")
	for i, addr := range brokerAddrs {
		fmt.Printf("    %d. %s\n", i+1, addr)
	}
	for {
		fmt.Printf("  Opção (1-%d): ", len(brokerAddrs))
		scanner.Scan()
		idx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err == nil && idx >= 1 && idx <= len(brokerAddrs) {
			brokerGateway = brokerAddrs[idx-1]
			fmt.Printf("  [REDE] Conexão estabelecida com o gateway primário: %s\n", brokerGateway)
			break
		}
		fmt.Println("  ✗ Opção inválida.")
	}

	// 4. MENU PRINCIPAL
	for {
		exibirMenu()
		fmt.Print("  Opção: ")
		if !scanner.Scan() {
			break
		}
		opcao := strings.TrimSpace(scanner.Text())

		switch opcao {
		case "1":
			enviarRequisicaoManual(scanner)
		case "2":
			enviarRequisicaoAleatoria()
		case "3":
			consultarFila()
		case "4":
			enviarMultiplasRequisicoes(scanner)
		case "5":
			consultarSaldoEfetivo()
		case "6":
			consultarLedger()
		case "7":
			enviarRequisicaoMaliciosa(scanner)
		case "0":
			fmt.Println("\n  Encerrando terminal do operador.")
			return
		default:
			fmt.Println("\n  ✗ Opção inválida.")
		}
	}
}

func exibirMenu() {
	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────┐")
	fmt.Printf("  │ EMPRESA: %-25s │\n", strings.ToUpper(nomeCompanhia))
	fmt.Printf("  │ NÓ ATUAL: %-24s │\n", brokerGateway)
	fmt.Println("  ├──────────────────────────────────────┤")
	fmt.Println("  │  1. Enviar requisição manual         │")
	fmt.Println("  │  2. Enviar requisição aleatória      │")
	fmt.Println("  │  3. Consultar fila (Via Gateway)     │")
	fmt.Println("  │  4. Enviar múltiplas requisições     │")
	fmt.Println("  │  5. Auditoria de Consenso Global     │")
	fmt.Println("  │  6. Consultar Histórico (Ledger)     │")
	fmt.Println("  │  7. Enviar requisição maliciosa      │")
	fmt.Println("  │  0. Sair                             │")
	fmt.Println("  └──────────────────────────────────────┘")
}

func enviarRequisicaoManual(scanner *bufio.Scanner) {
	fmt.Println()
	fmt.Println("  ── NOVA REQUISIÇÃO (DADOS GEOGRÁFICOS) ─────────────────────")

	fmt.Println("  Tipos de alerta:")
	alertas := []string{
		"1. Embarcação à deriva detectada",
		"2. Objeto não identificado no canal",
		"3. Bloqueio parcial de rota comercial",
		"4. Sinal de socorro (SOS) captado",
		"5. Mancha de óleo identificada",
	}
	for _, a := range alertas {
		fmt.Printf("     %s\n", a)
	}

	fmt.Print("  Escolha o tipo (1-5): ")
	scanner.Scan()
	tipo, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || tipo < 1 || tipo > 5 {
		fmt.Println("  ✗ Tipo inválido.")
		return
	}
	descricoes := []string{
		"Embarcação à deriva detectada",
		"Objeto não identificado no canal",
		"Bloqueio parcial de rota comercial",
		"Sinal de socorro (SOS) captado",
		"Mancha de óleo identificada",
	}

	fmt.Print("  Prioridade (1=baixa … 5=crítica): ")
	scanner.Scan()
	prioridade, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || prioridade < 1 || prioridade > 5 {
		fmt.Println("  ✗ Prioridade inválida.")
		return
	}

	// O Setor define para onde o Drone vai voar (1 a 9 zonas mapeadas)
	fmt.Print("  Coordenada Geográfica / Setor afetado (1 a 9): ")
	scanner.Scan()
	setor, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || setor < 1 || setor > 9 {
		fmt.Println("  ✗ Setor inválido.")
		return
	}

	custoMissao := float64(prioridade * 5)
	txAssinada := criarTransacaoAssinada(custoMissao, setor)

	req := models.Requisicao{
		ID:         fmt.Sprintf("REQ-%d-%04d", setor, mathrand.Intn(10000)),
		Setor:      setor,
		Prioridade: prioridade,
		Descricao:  descricoes[tipo-1],
		Status:     models.StatusPendente,
		CreatedAt:  time.Now(),
		Transacao:  txAssinada,
	}

	enviarComFailover(req)
}

func enviarRequisicaoAleatoria() {
	descricoes := []string{
		"Embarcação à deriva detectada",
		"Objeto não identificado no canal",
		"Bloqueio parcial de rota comercial",
		"Sinal de socorro (SOS) captado",
		"Mancha de óleo identificada",
	}

	setorGeografico := mathrand.Intn(9) + 1 // Gera de 1 a 9
	prioridade := mathrand.Intn(5) + 1
	custoMissao := float64(prioridade * 5)

	txAssinada := criarTransacaoAssinada(custoMissao, setorGeografico)

	req := models.Requisicao{
		ID:         fmt.Sprintf("REQ-%d-%04d", setorGeografico, mathrand.Intn(10000)),
		Setor:      setorGeografico,
		Prioridade: prioridade,
		Descricao:  descricoes[mathrand.Intn(len(descricoes))],
		Status:     models.StatusPendente,
		CreatedAt:  time.Now(),
		Transacao:  txAssinada,
	}

	enviarComFailover(req)
}

func enviarMultiplasRequisicoes(scanner *bufio.Scanner) {
	fmt.Print("\n  Quantas requisições enviar? ")
	scanner.Scan()
	n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if err != nil || n < 1 {
		fmt.Println("  ✗ Quantidade inválida.")
		return
	}
	fmt.Printf("\n  Enviando %d requisições via %s...\n", n, brokerGateway)
	for i := 0; i < n; i++ {
		enviarRequisicaoAleatoria()
		time.Sleep(300 * time.Millisecond)
	}
	fmt.Printf("  ✓ %d requisições enviadas.\n", n)
}

func extrairBrokerID(addr string) int {
	partes := strings.Split(addr, ":")
	if len(partes) > 0 {
		nome := strings.Split(partes[0], "-")
		if len(nome) == 2 {
			id, _ := strconv.Atoi(nome[1])
			return id
		}
	}
	return 0
}

func enviarComFailover(req models.Requisicao) {
	prioLabel := strings.Repeat("★", req.Prioridade) + strings.Repeat("☆", 5-req.Prioridade)
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("  │ ENVIANDO: %-48s│\n", req.Descricao)
	fmt.Printf("  │ ID da Missão: %-48s│\n", req.ID)
	fmt.Printf("  │ Localização: Setor %-2d  Prioridade: %s                   │\n", req.Setor, prioLabel)
	fmt.Println("  └─────────────────────────────────────────────────────────┘")

	type Nodo struct {
		ID   int
		Addr string
	}
	var nodos []Nodo
	for _, a := range brokerAddrs {
		nodos = append(nodos, Nodo{ID: extrairBrokerID(a), Addr: a})
	}

	sort.Slice(nodos, func(i, j int) bool {
		return nodos[i].ID > nodos[j].ID
	})

	idxGateway := 0
	for i, n := range nodos {
		if n.Addr == brokerGateway {
			idxGateway = i
			break
		}
	}

	alvosDeEnvio := []string{brokerGateway}
	for i := 1; i < len(nodos); i++ {
		proximoIdx := (idxGateway + i) % len(nodos)
		alvosDeEnvio = append(alvosDeEnvio, nodos[proximoIdx].Addr)
	}

	for i, addr := range alvosDeEnvio {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			if i == 0 {
				fmt.Printf("  [!] Gateway Primário (%s) offline! Iniciando FAILOVER (Ordem Decrescente)...\n", addr)
			} else {
				fmt.Printf("  [!] Nó de backup %s offline. Procurando próximo menor...\n", addr)
			}
			continue
		}

		envelope := models.MensagemDistribuida{
			Tipo:      models.MsgSyncNew,
			SenderID:  req.Setor,
			Timestamp: 0,
			Payload:   req,
		}

		if err := json.NewEncoder(conn).Encode(envelope); err != nil {
			conn.Close()
			continue
		}
		conn.Close()

		if i == 0 {
			fmt.Printf("  ✓ Requisição entregue com sucesso ao Gateway Oficial (%s)\n", addr)
		} else {
			fmt.Printf("  ✓ [FAILOVER SALVO] Requisição assumida pelo Nó de Backup (%s)\n", addr)
		}
		return
	}
	fmt.Println("  ✗ ERRO CRÍTICO: Toda a malha P2P encontra-se offline!")
}

// consultarSaldoEfetivo é uma função exclusiva de "Raio-X" (Auditoria) para
// validar se o Consenso Distribuído da Blockchain está a funcionar corretamente.
func consultarSaldoEfetivo() {
	fmt.Println()
	fmt.Println("  ╔════ AUDITORIA DE CONSENSO DISTRIBUÍDO (RAIO-X DA REDE) ════╗")
	fmt.Printf("  ║ Conta: %-51s ║\n", companhiaPubKey[:50]+"...")
	fmt.Println("  ╚════════════════════════════════════════════════════════════╝")

	for _, addr := range brokerAddrs {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			fmt.Printf("    ✗ %-18s -> [FALHA DE COMUNICAÇÃO - OFFLINE]\n", addr)
			continue
		}

		msg := models.MensagemDistribuida{
			Tipo:      models.MsgConsultaSaldo,
			SenderID:  0,
			Timestamp: time.Now().UnixMilli(),
			Payload:   companhiaPubKey,
		}

		if err := json.NewEncoder(conn).Encode(msg); err != nil {
			conn.Close()
			continue
		}

		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		var saldo float64
		if err := json.NewDecoder(conn).Decode(&saldo); err != nil {
			fmt.Printf("    ✗ %-18s -> [ERRO NA DESCODIFICAÇÃO DO SALDO]\n", addr)
		} else {
			fmt.Printf("    ✓ %-18s -> Saldo Reconhecido: %.2f créditos\n", addr, saldo)
		}
		conn.Close()
	}
}

// consultarFila é o comportamento real de um cliente: só interroga o seu Gateway Oficial.
func consultarFila() {
	fmt.Println()
	fmt.Printf("  Consultando estado global estritamente através do gateway %s...\n", brokerGateway)

	conn, err := net.DialTimeout("tcp", brokerGateway, 2*time.Second)
	if err != nil {
		fmt.Printf("  ✗ Gateway %s indisponível para leitura.\n", brokerGateway)
		return
	}
	defer conn.Close()

	msg := models.MensagemDistribuida{
		Tipo:      models.MsgConsultaFila,
		SenderID:  0,
		Timestamp: 0,
	}

	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		fmt.Println("  ✗ Falha ao enviar requisição.")
		return
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var resposta models.RespostaFila
	if err := json.NewDecoder(conn).Decode(&resposta); err != nil {
		fmt.Println("  ✗ Falha ao decodificar a estrutura de resposta do broker.")
		return
	}

	exibirFila(resposta.Requisicoes, brokerGateway)
}

func exibirFila(lista []models.Requisicao, broker string) {
	fmt.Printf("\n  ╔══ FILA DE REQUISIÇÕES — Lida a partir do nó %s (%d missões) ══╗\n", broker, len(lista))

	grupos := map[models.StatusRequisicao][]models.Requisicao{
		models.StatusPendente:      {},
		models.StatusEmAtendimento: {},
		models.StatusConcluido:     {},
	}
	for _, r := range lista {
		grupos[r.Status] = append(grupos[r.Status], r)
	}

	ordemStatus := []models.StatusRequisicao{
		models.StatusEmAtendimento,
		models.StatusPendente,
		models.StatusConcluido,
	}
	icones := map[models.StatusRequisicao]string{
		models.StatusEmAtendimento: "🚁",
		models.StatusPendente:      "⏳",
		models.StatusConcluido:     "✓",
	}

	for _, status := range ordemStatus {
		reqs := grupos[status]
		if len(reqs) == 0 {
			continue
		}
		fmt.Printf("\n  %s %s (%d)\n", icones[status], status, len(reqs))
		fmt.Println("  ─────────────────────────────────────────────────────────")
		for _, r := range reqs {
			prioLabel := strings.Repeat("★", r.Prioridade) + strings.Repeat("☆", 5-r.Prioridade)
			drone := r.DroneID
			if drone == "" {
				drone = "—"
			}
			fmt.Printf("  │ %-12s │ %s │ Setor %d │ Drone: %-6s │ %s\n",
				r.ID, prioLabel, r.Setor, drone, r.Descricao)
		}
	}

	if len(lista) == 0 {
		fmt.Println("  │ Fila vazia — nenhuma requisição processada ou pendente no cluster.")
	}
	fmt.Println("\n  ╚══════════════════════════════════════════════════════════════════════════╝")
}

func inicializarIdentidadePersistente() {
	arquivoCarteira := fmt.Sprintf("data/carteira_%s.pem", nomeCompanhia)

	if _, err := os.Stat(arquivoCarteira); err == nil {
		pemBytes, err := ioutil.ReadFile(arquivoCarteira)
		if err == nil {
			bloco, _ := pem.Decode(pemBytes)
			if bloco != nil {
				chave, err := x509.ParseECPrivateKey(bloco.Bytes)
				if err == nil {
					chavePrivada = chave
					pubKeyBytes := elliptic.Marshal(elliptic.P256(), chavePrivada.PublicKey.X, chavePrivada.PublicKey.Y)
					companhiaPubKey = hex.EncodeToString(pubKeyBytes)
					fmt.Printf("\n  [SEC] Carteira da '%s' encontrada no disco!\n  Chave Pública Autenticada:\n  %s...\n", strings.ToUpper(nomeCompanhia), companhiaPubKey[:40])
					return
				}
			}
		}
	}

	var err error
	chavePrivada, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic("Falha ao gerar chaves criptográficas: " + err.Error())
	}

	bytesChave, _ := x509.MarshalECPrivateKey(chavePrivada)
	blocoPem := &pem.Block{Type: "EC PRIVATE KEY", Bytes: bytesChave}
	_ = ioutil.WriteFile(arquivoCarteira, pem.EncodeToMemory(blocoPem), 0644)

	pubKeyBytes := elliptic.Marshal(elliptic.P256(), chavePrivada.PublicKey.X, chavePrivada.PublicKey.Y)
	companhiaPubKey = hex.EncodeToString(pubKeyBytes)

	fmt.Printf("\n  [SEC] Nova Carteira criada para '%s' e guardada no disco.\n  Chave Pública:\n  %s...\n", strings.ToUpper(nomeCompanhia), companhiaPubKey[:40])
}

func criarTransacaoAssinada(valor float64, setor int) models.TransacaoToken { // <-- ADD setor
	tx := models.TransacaoToken{
		CompanhiaPubKey: companhiaPubKey,
		Valor:           valor,
		Timestamp:       time.Now().UnixMilli(),
	}

	hash := tx.GerarHashDados(setor) // <-- PASSA O SETOR PARA O HASH
	// ... o resto da função continua igual ...
	assinaturaBytes, err := ecdsa.SignASN1(rand.Reader, chavePrivada, hash)
	if err != nil {
		panic("Falha ao assinar transação: " + err.Error())
	}

	tx.Assinatura = hex.EncodeToString(assinaturaBytes)
	return tx
}

// consultarLedger faz o download dos blocos imutáveis guardados no disco do Broker.
func consultarLedger() {
	fmt.Println()
	fmt.Printf("  Descarregando cópia do Ledger Imutável via %s...\n", brokerGateway)

	conn, err := net.DialTimeout("tcp", brokerGateway, 2*time.Second)
	if err != nil {
		fmt.Printf("  ✗ Gateway %s indisponível para leitura do disco.\n", brokerGateway)
		return
	}
	defer conn.Close()

	msg := models.MensagemDistribuida{
		Tipo:      models.MsgConsultaLedger,
		SenderID:  0,
		Timestamp: 0,
	}

	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		fmt.Println("  ✗ Falha ao solicitar os dados do Ledger.")
		return
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var cadeia []models.Bloco
	if err := json.NewDecoder(conn).Decode(&cadeia); err != nil {
		fmt.Println("  ✗ Falha ao decodificar a estrutura da Blockchain.")
		return
	}

	fmt.Printf("\n  ╔══ BLOCKCHAIN (LEDGER PERSISTENTE) — %s (%d Blocos Selados) ══╗\n", brokerGateway, len(cadeia))
	if len(cadeia) == 0 {
		fmt.Println("  │ Ledger vazio. Nenhuma missão foi concluída e selada ainda.")
	} else {
		for _, bloco := range cadeia {
			fmt.Printf("  │ 🧱 Bloco #%-4d │ Missão: %-14s │ Drone: %-6s\n", bloco.Index, bloco.MissaoID, bloco.Laudo.DroneID)
			fmt.Printf("  │    Hash: %s...\n", bloco.HashAtual[:32])
			fmt.Printf("  │    Custo: %.2f créditos (Pago por Cia: %s...)\n", bloco.Transacao.Valor, bloco.Transacao.CompanhiaPubKey[:8])
			fmt.Println("  ├─────────────────────────────────────────────────────────")
		}
	}
	fmt.Println("  ╚══════════════════════════════════════════════════════════════════════════╝")
}

func enviarRequisicaoMaliciosa(scanner *bufio.Scanner) {
	fmt.Println("\n  ┌── MENU DE ATAQUES ──┐")
	fmt.Println("  1. Adulterar Valor (Ataque de Integridade)")
	fmt.Println("  2. Adulterar Setor (Ataque de Rota)")
	fmt.Println("  3. Replay Attack (Enviar pacote duplicado)")
	fmt.Print("  Opção: ")
	scanner.Scan()
	opcao := scanner.Text()

	tx := criarTransacaoAssinada(10., 1)
	req := models.Requisicao{
		ID:    "ATAQUE-" + strconv.Itoa(mathrand.Intn(1000)),
		Setor: 1, Prioridade: 1, Descricao: "FRAUDE", Status: models.StatusPendente,
		Transacao: tx,
	}

	switch opcao {
	case "1":
		req.Transacao.Valor = 999.0 // Fraude: muda valor
	case "2":
		req.Setor = 9 // Fraude: muda destino sem refazer assinatura
	case "3":
		// Envia o mesmo objeto duas vezes
		enviarComFailover(req)
	}

	fmt.Println("  [!] Enviando pacote fraudulento...")
	enviarComFailover(req)
}
