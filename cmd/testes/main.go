package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"net"
	"ormuz_distribuido/internal/models"
	"os"
	"strings"
	"time"
)

var (
	brokerAddrs     []string
	chavePrivada    *ecdsa.PrivateKey
	companhiaPubKey string
)

type resultado struct {
	nome    string
	passou  bool
	detalhe string
}

func main() {
	addrsRaw := os.Getenv("BROKER_ADDRS")
	if addrsRaw == "" {
		addrsRaw = "broker-1:9000,broker-2:9000,broker-3:9000,broker-4:9000"
	}
	for _, a := range strings.Split(addrsRaw, ",") {
		brokerAddrs = append(brokerAddrs, strings.TrimSpace(a))
	}

	mathrand.Seed(time.Now().UnixNano())
	gerarCarteiraDeTestes() // Gera uma identidade criptográfica para os testes

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     SUÍTE DE TESTES DE SEGURANÇA E CONSENSO — ORMUZ          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	resultados := []resultado{}

	resultados = append(resultados, testeConectividadeBrokers())
	resultados = append(resultados, testeEnvioRequisicaoLegitima())
	resultados = append(resultados, testeAtaqueIntegridade())
	resultados = append(resultados, testeReplayAttack())
	resultados = append(resultados, testeConsensoSaldos())
	resultados = append(resultados, testeIntegridadeBlockchain())

	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════")
	fmt.Println("  RELATÓRIO DE AUDITORIA FINAL")
	fmt.Println("══════════════════════════════════════════════════════════════")
	passou := 0
	for _, r := range resultados {
		icone := "✓"
		if !r.passou {
			icone = "✗"
		} else {
			passou++
		}
		fmt.Printf("  %s %-45s %s\n", icone, r.nome, r.detalhe)
	}
	fmt.Printf("\n  Resultado: %d/%d testes passaram de forma impecável\n", passou, len(resultados))
	fmt.Println("══════════════════════════════════════════════════════════════")
}

// =================================================================
// SEGURANÇA E CRIPTOGRAFIA PARA OS TESTES
// =================================================================
func gerarCarteiraDeTestes() {
	chavePrivada, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubKeyBytes := elliptic.Marshal(elliptic.P256(), chavePrivada.PublicKey.X, chavePrivada.PublicKey.Y)
	companhiaPubKey = hex.EncodeToString(pubKeyBytes)
}

func assinarTransacao(valor float64, setor int) models.TransacaoToken {
	tx := models.TransacaoToken{
		CompanhiaPubKey: companhiaPubKey,
		Valor:           valor,
		Timestamp:       time.Now().UnixMilli(),
	}
	payload := fmt.Sprintf("%s:%.2f:%d:%d", tx.CompanhiaPubKey, tx.Valor, setor, tx.Timestamp)
	h := sha256.Sum256([]byte(payload))
	assinaturaBytes, _ := ecdsa.SignASN1(rand.Reader, chavePrivada, h[:])
	tx.Assinatura = hex.EncodeToString(assinaturaBytes)
	return tx
}

func novaRequisicaoAssinada(setor, prioridade int, descricao string) models.Requisicao {
	custo := float64(prioridade * 5)
	return models.Requisicao{
		ID:         fmt.Sprintf("TESTE-%d-%04d", setor, mathrand.Intn(10000)),
		Setor:      setor,
		Prioridade: prioridade,
		Descricao:  descricao,
		Status:     models.StatusPendente,
		CreatedAt:  time.Now(),
		Transacao:  assinarTransacao(custo, setor),
	}
}

// =================================================================
// OS CASOS DE TESTE
// =================================================================

func testeConectividadeBrokers() resultado {
	nome := "Conectividade com a malha P2P"
	conectados := 0
	for _, addr := range brokerAddrs {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			conectados++
		}
	}
	if conectados == 0 {
		return resultado{nome, false, "Nenhum broker acessível"}
	}
	return resultado{nome, true, fmt.Sprintf("%d/%d brokers operacionais", conectados, len(brokerAddrs))}
}

func testeEnvioRequisicaoLegitima() resultado {
	nome := "Envio Criptografado Legítimo (ECDSA)"
	req := novaRequisicaoAssinada(1, 3, "Missão de Teste Legítima")

	if err := enviarRequisicao(req); err != nil {
		return resultado{nome, false, "Falha ao enviar."}
	}
	time.Sleep(2 * time.Second) // Tempo para propagar na malha

	fila, err := consultarFila(brokerAddrs[0])
	if err != nil {
		return resultado{nome, false, "Falha ao consultar Mempool."}
	}

	for _, r := range fila {
		if r.ID == req.ID {
			return resultado{nome, true, "Assinatura validada e aceite na Mempool"}
		}
	}
	return resultado{nome, false, "Requisição legítima foi descartada"}
}

func testeAtaqueIntegridade() resultado {
	nome := "Defesa contra Adulteração de Rota (Setor)"

	// Cria uma requisição legítima para o Setor 1
	reqMaliciosa := novaRequisicaoAssinada(1, 3, "Ataque de Rota")

	// HACK: Altera o setor para 9 em trânsito (quebra o Hash da Assinatura)
	reqMaliciosa.Setor = 9

	enviarRequisicao(reqMaliciosa)
	time.Sleep(1 * time.Second)

	fila, _ := consultarFila(brokerAddrs[0])
	for _, r := range fila {
		if r.ID == reqMaliciosa.ID {
			return resultado{nome, false, "FALHA CRÍTICA: O broker aceitou uma rota adulterada!"}
		}
	}
	return resultado{nome, true, "Broker detetou quebra de Hash e rejeitou o pacote"}
}

func testeReplayAttack() resultado {
	nome := "Defesa contra Replay Attack"
	req := novaRequisicaoAssinada(2, 2, "Alvo de Replay Attack")

	// Envia a mesma requisição exata DUAS VEZES (Looping de Fraude)
	enviarRequisicao(req)
	enviarRequisicao(req)

	time.Sleep(2 * time.Second)

	fila, _ := consultarFila(brokerAddrs[0])
	ocorrencias := 0
	for _, r := range fila {
		if r.ID == req.ID {
			ocorrencias++
		}
	}

	if ocorrencias == 1 {
		return resultado{nome, true, "Cópia exata bloqueada. Unicidade garantida."}
	} else if ocorrencias > 1 {
		return resultado{nome, false, fmt.Sprintf("FALHA: Pacote duplicado %d vezes!", ocorrencias)}
	}
	return resultado{nome, false, "O pacote original também foi rejeitado"}
}

func testeConsensoSaldos() resultado {
	nome := "Consenso Global de Saldos (Duplo Gasto)"
	time.Sleep(1 * time.Second) // Aguarda a Mempool estabilizar

	var saldoAnterior float64 = -1

	for _, addr := range brokerAddrs {
		saldoAtual, err := consultarSaldo(addr, companhiaPubKey)
		if err != nil {
			return resultado{nome, false, "Falha na auditoria de " + addr}
		}
		if saldoAnterior != -1 && saldoAtual != saldoAnterior {
			return resultado{nome, false, "Desincronização financeira (Forks locais detetados!)"}
		}
		saldoAnterior = saldoAtual
	}
	return resultado{nome, true, fmt.Sprintf("Todos os nós concordam: %.2f Créditos", saldoAnterior)}
}

func testeIntegridadeBlockchain() resultado {
	nome := "Integridade Criptográfica do Ledger (Hashes)"
	time.Sleep(1 * time.Second)

	cadeia, err := consultarLedgerCompleto(brokerAddrs[0])
	if err != nil {
		return resultado{nome, false, "Falha ao baixar Blockchain"}
	}

	if len(cadeia) == 0 {
		return resultado{nome, true, "Ledger válido (Atualmente Vazio)"}
	}

	hashEsperado := "0000000000000000000000000000000000000000000000000000000000000000"
	for _, bloco := range cadeia {
		if bloco.HashAnterior != hashEsperado {
			return resultado{nome, false, fmt.Sprintf("Corrente quebrada no Bloco %d!", bloco.Index)}
		}

		// Recalcula o Hash na hora para garantir que ninguém mexeu no disco local
		hashRecalculado := bloco.CalcularHash()
		if hashRecalculado != bloco.HashAtual {
			return resultado{nome, false, fmt.Sprintf("Adulteração local detetada no Bloco %d!", bloco.Index)}
		}
		hashEsperado = bloco.HashAtual
	}

	return resultado{nome, true, fmt.Sprintf("Cadeia Inviolada (%d blocos validados)", len(cadeia))}
}

// =================================================================
// FUNÇÕES AUXILIARES DE REDE
// =================================================================
func enviarRequisicao(req models.Requisicao) error {
	for _, addr := range brokerAddrs {
		if err := enviarRequisicaoParaAddr(addr, req); err == nil {
			return nil
		}
	}
	return fmt.Errorf("rede P2P inacessível")
}

func enviarRequisicaoParaAddr(addr string, req models.Requisicao) error {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return json.NewEncoder(conn).Encode(models.MensagemDistribuida{
		Tipo:      models.MsgSyncNew,
		SenderID:  req.Setor,
		Timestamp: 0,
		Payload:   req,
	})
}

func consultarFila(addr string) ([]models.Requisicao, error) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	json.NewEncoder(conn).Encode(models.MensagemDistribuida{Tipo: models.MsgConsultaFila})
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var resposta models.RespostaFila
	if err := json.NewDecoder(conn).Decode(&resposta); err != nil {
		return nil, err
	}
	return resposta.Requisicoes, nil
}

func consultarSaldo(addr, pubKey string) (float64, error) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	json.NewEncoder(conn).Encode(models.MensagemDistribuida{Tipo: models.MsgConsultaSaldo, Payload: pubKey})
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var saldo float64
	err = json.NewDecoder(conn).Decode(&saldo)
	return saldo, err
}

func consultarLedgerCompleto(addr string) ([]models.Bloco, error) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	json.NewEncoder(conn).Encode(models.MensagemDistribuida{Tipo: models.MsgConsultaLedger})
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var cadeia []models.Bloco
	err = json.NewDecoder(conn).Decode(&cadeia)
	return cadeia, err
}
