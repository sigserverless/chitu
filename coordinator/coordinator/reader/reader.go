package reader

type Reader struct {
	IP string `json:"ip"`
}

func NewReader(ip string) *Reader {
	return &Reader{
		IP: ip,
	}
}
