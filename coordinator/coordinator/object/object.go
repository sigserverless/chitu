package object

import (
	"coordinator/reader"
	"coordinator/writer"
)

type Object struct {
	Key     string           `json:"key"`
	Writers []*writer.Writer `json:"writers"`
	Readers []*reader.Reader `json:"readers"`
	DagID   string           `json:"dagId"`
}

func NewObjectWithReader(reader1 *reader.Reader, key, dagID string) *Object {
	return &Object{
		Key:     key,
		Writers: nil,
		Readers: []*reader.Reader{reader1},
		DagID:   dagID,
	}
}

func NewObjectWithWriter(writer1 *writer.Writer, key, dagID string) *Object {
	return &Object{
		Key:     key,
		Writers: []*writer.Writer{writer1},
		Readers: nil,
		DagID:   dagID,
	}
}

func contains(slice []*reader.Reader, element *reader.Reader) bool {
	for _, item := range slice {
		if item.IP == element.IP {
			return true
		}
	}
	return false
}

func (o *Object) AddReader(reader1 *reader.Reader) {
	if o.Readers == nil {
		o.Readers = make([]*reader.Reader, 0)
	}

	if !contains(o.Readers, reader1) {
		o.Readers = append(o.Readers, reader1)
	}
}

func (o *Object) AddWriter(writer1 *writer.Writer) {
	if o.Writers == nil {
		o.Writers = make([]*writer.Writer, 0)
	}
	o.Writers = append(o.Writers, writer1)
}

func (o *Object) GetReaders() []*reader.Reader {
	return o.Readers
}

func (o *Object) GetWriters() []*writer.Writer {
	return o.Writers
}
