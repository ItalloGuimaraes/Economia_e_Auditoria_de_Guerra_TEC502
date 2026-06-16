package models

// TipoMensagem identifica o propósito de cada mensagem trocada na malha P2P.
type TipoMensagem string

const (
	MsgJoin           TipoMensagem = "JOIN"
	MsgJoinACK        TipoMensagem = "JOIN_ACK"
	MsgSyncNew        TipoMensagem = "SYNC_NEW"
	MsgSyncUpdate     TipoMensagem = "SYNC_UPDATE"
	MsgFullSync       TipoMensagem = "FULL_SYNC"
	MsgReqDrone       TipoMensagem = "REQ_DRONE"
	MsgReplyOK        TipoMensagem = "REPLY_OK"
	MsgDroneHeartbeat TipoMensagem = "DRONE_HEARTBEAT"
	MsgDroneConcluido TipoMensagem = "DRONE_CONCLUIDO"
	MsgConsultaFila   TipoMensagem = "CONSULTA_FILA"
	MsgNovoBloco      TipoMensagem = "NOVO_BLOCO"
	MsgConsultaSaldo  TipoMensagem = "CONSULTA_SALDO"
	MsgConsultaLedger TipoMensagem = "CONSULTA_LEDGER"
)

// MensagemDistribuida é o envelope JSON de toda comunicação TCP do sistema.
type MensagemDistribuida struct {
	Tipo      TipoMensagem `json:"tipo"`
	SenderID  int          `json:"sender_id"`
	Timestamp int64        `json:"timestamp"`
	Payload   interface{}  `json:"payload"`
}

// JoinRequest é o payload de MsgJoin e MsgJoinACK.
// Carrega o ID e o endereço externamente acessível do broker.
type JoinRequest struct {
	ID   int    `json:"id"`
	Addr string `json:"addr"`
}

// DroneStatus é o payload de MsgDroneHeartbeat e MsgDroneConcluido.
type DroneStatus struct {
	DroneID   string `json:"drone_id"`
	MissionID string `json:"mission_id"`
}

// RespostaFila é o payload de resposta ao MsgConsultaFila.
type RespostaFila struct {
	Requisicoes []Requisicao `json:"requisicoes"`
}
