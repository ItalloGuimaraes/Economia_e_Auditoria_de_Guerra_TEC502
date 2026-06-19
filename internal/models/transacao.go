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
func (t *TransacaoToken) GerarHashDados(setor int) []byte {
	payload := fmt.Sprintf("%s:%.2f:%d:%d", t.CompanhiaPubKey, t.Valor, setor, t.Timestamp)
	h := sha256.Sum256([]byte(payload))
	return h[:]
}
