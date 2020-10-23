package model

import (
	"fmt"
	"sort"
)

// ErrThreadIDNotFound represent a missing thread id error
type ErrThreadIDNotFound struct {
	tid string
}

// Error satisfies the error inteface
func (e ErrThreadIDNotFound) Error() string {
	return fmt.Sprintf("Thread ID %s not found", e.tid)
}

// ErrMessageIDNotFound represent a missing message id error
type ErrMessageIDNotFound struct {
	tid string
}

// Error satisfies the error inteface
func (e ErrMessageIDNotFound) Error() string {
	return fmt.Sprintf("Message ID %s not found", e.tid)
}

// Message represents a message in the threader app
type Message struct {
	ID        string `json:"id"`
	ThreadID  string `json:"thread_id"`
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
	Votes     int    `json:"votes"`
}

// Thread represents a thread in the threader app
type Thread struct {
	ID       string     `json:"id"`
	Topic    string     `json:"topic"`
	Messages []*Message `json:"messages"`
}

// Objects represents a simulated persistent memory for our data objects
type Objects struct {
	threads map[string]*Thread
}

// NewObjectsStore returns an empty new Objects store
func NewObjectsStore() *Objects {
	return &Objects{}
}

// Init initializes the store
func (s *Objects) Init() error {
	s.threads = make(map[string]*Thread)
	return nil
}

// Close frees all resources used by te store
func (s *Objects) Close() {}

// Threads return the list of stored threads
func (s *Objects) Threads() []Thread {
	ths := make([]Thread, 0, len(s.threads))
	for _, t := range s.threads {
		ths = append(ths, *t)
	}

	return ths
}

// GetThread returns a reference to the stored thread with ThreadID tid
func (s *Objects) GetThread(tid string) (*Thread, error) {
	t, ok := s.threads[tid]
	if !ok {
		return nil, ErrThreadIDNotFound{tid}
	}

	return t, nil
}

// GetMessage returns a reference to the stored message with ThreadID tid and MessageID mid
func (s *Objects) GetMessage(tid, mid string) (*Message, error) {
	t, err := s.GetThread(tid)
	if err != nil {
		return nil, err
	}

	for _, m := range t.Messages {
		if m.ID == mid {
			return m, nil
		}
	}

	return nil, ErrMessageIDNotFound{mid}
}

// AddThread adds the thread t to the list of stored threads
func (s *Objects) AddThread(t *Thread) {
	s.threads[t.ID] = t
}

// AddMessage add the message m to the thread with id tid
// it returns an erorr if tid does not exist
func (s *Objects) AddMessage(tid string, m *Message) error {
	thread, ok := s.threads[tid]
	if !ok {
		return ErrThreadIDNotFound{tid}
	}

	thread.Messages = append(thread.Messages, m)
	sort.Slice(thread.Messages, func(i, j int) bool {
		return thread.Messages[i].Timestamp < thread.Messages[j].Timestamp
	})

	return nil
}
