package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// LaudoAuditoria armazena o resultado da missão, tornando-o à prova de adulteração.
type LaudoAuditoria struct {
	DroneID   string `json:"drone_id"`
	Relatorio string `json:"relatorio"` // Ex: "Rota segura", "Bloqueio detectado"
}

// Bloco é a unidade fundamental da Blockchain.
// Encapsula a transação financeira e o laudo operacional da missão concluída.
type Bloco struct {
	Index        int            `json:"index"`
	Timestamp    int64          `json:"timestamp"`
	MissaoID     string         `json:"missao_id"`
	Transacao    TransacaoToken `json:"transacao"`
	Laudo        LaudoAuditoria `json:"laudo"`
	HashAnterior string         `json:"hash_anterior"` // Elo com o bloco anterior
	HashAtual    string         `json:"hash_atual"`    // Selo de integridade deste bloco
}

// CalcularHash gera o SHA-256 do bloco baseando-se em todo o seu conteúdo.
// Qualquer alteração nos dados do laudo ou da transação muda o hash,
// tornando a adulteração detectável pela cadeia.
//
// NOTA: HashAtual é excluído do cálculo (não faz parte do record) porque
// ele é o próprio resultado desta função — incluí-lo seria circularidade.
func (b *Bloco) CalcularHash() string {
	record := fmt.Sprintf("%d%d%s%s%.2f%d%s%s%s%s",
		b.Index, b.Timestamp, b.MissaoID,
		b.Transacao.CompanhiaPubKey, b.Transacao.Valor, b.Transacao.Timestamp, b.Transacao.Assinatura,
		b.Laudo.DroneID, b.Laudo.Relatorio, b.HashAnterior)

	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// Blockchain encapsula a cadeia de blocos no estado local do Broker.
type Blockchain struct {
	Cadeia []Bloco `json:"cadeia"`
}
