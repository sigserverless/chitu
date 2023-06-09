package protocol

type InvokeReq struct {
	Req   string `json:"req"`
	DagId string `json:"dagId"`
	// Key   string `json:"key"`
}

type InvokeRes struct {
	Res   []byte `json:"res"`
	InvId string `json:"invId"`
}

type GetNotifyReq struct {
	Key   string `json:"key"`
	InvId string `json:"invId"`
	DagId string `json:"dagId"`
	IP    string `json:"ip"`
}

type UpdateReq struct {
	Key   string   `json:"key"`
	DagId string   `json:"dagId"`
	Wops  [][]byte `json:"wop"`
	From  int      `json:"from"`
	// TODO: Since To is always equal to From + len(Wops), it
	// might not nned to be sent
	To int `json:"to"`
}

// type UpdateRes struct {
// 	Last int `json:"last"`
// }

type EndReq struct {
	Key   string `json:"key"`
	DagId string `json:"dagId"`
}

type CoordPutReq struct {
	Key   string `json:"key"`
	InvId string `json:"invId"`
	DagId string `json:"dagId"`
}

type CoordPutRes struct {
	Ips    []string `json:"ips"`
	InvIds []string `json:"invIds"`
}

type CoordGetReq struct {
	Key   string `json:"key"`
	InvId string `json:"invId"`
	DagId string `json:"dagId"`
}

type CoordTriggerReq struct {
	DagId string `json:"dagId"`
	Fname string `json:"fname"`
	Args  string `json:"args"`
}

const COORDINATOR_URL = "http://coordinator.openfaas-fn:8080"
