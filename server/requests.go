package server

type SubEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"d"`
}

type Request struct {
	OperationCode string      `json:"op"`
	Data          interface{} `json:"d"`
}
