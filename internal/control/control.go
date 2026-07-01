package control

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

var ErrNoGame = errors.New("no running goose game found — start one with `goose play`")

type Message struct {
	Label  string            `json:"label,omitempty"`
	Econ   []economy.Command `json:"econ,omitempty"`
	Events []events.Command  `json:"events,omitempty"`
}

func SocketPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "goose_control.sock"
	}
	return filepath.Join(home, ".goose", "control.sock")
}

type Server struct {
	ln   net.Listener
	path string
}

func Listen(handle func(Message)) (*Server, error) {
	path := SocketPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	_ = os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	s := &Server{ln: ln, path: path}
	go s.accept(handle)
	return s, nil
}

func (s *Server) accept(handle func(Message)) {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go func() {
			defer func() { _ = conn.Close() }()
			var msg Message
			if err := json.NewDecoder(conn).Decode(&msg); err == nil {
				handle(msg)
			}
		}()
	}
}

func (s *Server) Close() error {
	err := s.ln.Close()
	_ = os.Remove(s.path)
	return err
}

func Send(msg Message) error {
	conn, err := net.Dial("unix", SocketPath())
	if err != nil {
		return ErrNoGame
	}
	defer func() { _ = conn.Close() }()
	return json.NewEncoder(conn).Encode(msg)
}
