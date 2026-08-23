package archive

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

type Record struct {
	ID, Kind, RegionID    string
	Payload               any
	CreatedAt, ArchivedAt *time.Time
	Checksum              string
}
type Store struct {
	mu      sync.Mutex
	records map[string]Record
	blobs   map[string][]byte
}

func New() *Store { return &Store{records: map[string]Record{}, blobs: map[string][]byte{}} }
func (s *Store) Put(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" || r.Kind == "" {
		return errors.New("record identity required")
	}
	if _, ok := s.records[r.ID]; ok {
		return errors.New("record exists")
	}
	if r.CreatedAt == nil {
		n := time.Now().UTC()
		r.CreatedAt = &n
	}
	s.records[r.ID] = r
	return nil
}
func (s *Store) Archive(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.records[id]
	if !ok {
		return errors.New("record not found")
	}
	if r.ArchivedAt != nil {
		return nil
	}
	payload, e := json.Marshal(r.Payload)
	if e != nil {
		return e
	}
	var buf bytes.Buffer
	z := gzip.NewWriter(&buf)
	if _, e = z.Write(payload); e != nil {
		return e
	}
	if e = z.Close(); e != nil {
		return e
	}
	s.blobs[id] = buf.Bytes()
	now := time.Now().UTC()
	r.ArchivedAt = &now
	r.Checksum = fmtChecksum(buf.Bytes())
	r.Payload = nil
	s.records[id] = r
	return nil
}
func (s *Store) Restore(id string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.records[id]
	if !ok {
		return Record{}, errors.New("record not found")
	}
	if r.ArchivedAt == nil {
		return r, nil
	}
	blob := s.blobs[id]
	if fmtChecksum(blob) != r.Checksum {
		return Record{}, errors.New("archive checksum mismatch")
	}
	z, e := gzip.NewReader(bytes.NewReader(blob))
	if e != nil {
		return Record{}, e
	}
	data, e := io.ReadAll(z)
	_ = z.Close()
	if e != nil {
		return Record{}, e
	}
	if e = json.Unmarshal(data, &r.Payload); e != nil {
		return Record{}, e
	}
	return r, nil
}
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, id)
	delete(s.blobs, id)
	return nil
}
func (s *Store) List(kind string) []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Record{}
	for _, r := range s.records {
		if kind == "" || r.Kind == kind {
			out = append(out, r)
		}
	}
	return out
}
func fmtChecksum(data []byte) string {
	var sum uint64
	for _, b := range data {
		sum = sum*131 + uint64(b)
	}
	return fmt.Sprintf("%x", sum)
}
