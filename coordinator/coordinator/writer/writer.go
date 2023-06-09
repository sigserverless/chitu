package writer

type Writer struct {
	IP string `json:"ip"`
}

func NewWriter(ip string) *Writer {
	return &Writer{
		IP: ip,
	}
}
