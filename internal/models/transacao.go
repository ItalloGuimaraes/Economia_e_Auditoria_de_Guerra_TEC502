package models

import (
	"crypto/sha256"
	"fmt"
)

// TransacaoToken representa a transferência de créditos assinada digitalmente.
type TransacaoToken struct {
	CompanhiaPubKey string  `json:"companhia_pub_key"` // Identidade da empresa (Chave Pública Hex)
	Valor           float64 `json:"valor"`
	Timestamp       int64   `json:"timestamp"`  // Timestamp no momento da assinatura
	Assinatura      string  `json:"assinatura"` // Selo criptográfico ECDSA (Hex)
}

// GerarHashDados cria o resumo dos dados financeiros que será assinado pelo cliente
// e posteriormente verificado pelo Broker.
//
// IMPORTANTE: apenas CompanhiaPubKey, Valor e Timestamp entram no hash.
// O campo Assinatura é EXCLUÍDO deliberadamente para garantir que o hash
// calculado no cliente (antes de assinar) seja idêntico ao calculado no broker
// (quando Assinatura já está preenchida). Incluir Assinatura aqui causaria
// divergência de hash e faria toda transação legítima ser rejeitada como fraude.
func (t *TransacaoToken) GerarHashDados() []byte {
	// %.2f fixa o float em duas casas decimais, tornando a serialização determinística
	// independente de arredondamentos de ponto flutuante entre plataformas.
	payload := fmt.Sprintf("%s:%.2f:%d", t.CompanhiaPubKey, t.Valor, t.Timestamp)
	h := sha256.Sum256([]byte(payload))
	return h[:]
}
