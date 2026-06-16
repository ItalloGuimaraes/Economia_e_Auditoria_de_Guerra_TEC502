package models

import "time"

// StatusRequisicao representa o ciclo de vida de uma missão.
type StatusRequisicao string

const (
	StatusPendente      StatusRequisicao = "PENDENTE"
	StatusEmAtendimento StatusRequisicao = "EM_ATENDIMENTO"
	StatusConcluido     StatusRequisicao = "CONCLUIDO"
)

// Requisicao representa uma missão criada por um sensor/cliente e gerenciada pelos brokers.
type Requisicao struct {
	ID              string           `json:"id"`
	Setor           int              `json:"setor"`
	Prioridade      int              `json:"prioridade"`
	Descricao       string           `json:"descricao"`
	Status          StatusRequisicao `json:"status"`
	BrokerID        int              `json:"broker_id"`
	DroneID         string           `json:"drone_id"`
	Timestamp       int64            `json:"timestamp"`
	CreatedAt       time.Time        `json:"created_at"`
	IniciadoEm      time.Time        `json:"iniciado_em"`
	UltimoHeartbeat time.Time        `json:"ultimo_heartbeat"`
	Transacao       TransacaoToken   `json:"transacao"`
}
