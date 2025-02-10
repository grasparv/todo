package sse

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

type Session struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	logger *slog.Logger
	events chan []byte
	writer http.ResponseWriter
}

func newSession(ctx context.Context, logger *slog.Logger, w http.ResponseWriter) (*Session, error) {
	ctx, cancel := context.WithCancelCause(ctx)

	closeNotifier, ok := w.(http.CloseNotifier)
	if !ok {
		return nil, errors.New("close notification not supported")
	}

	closeNotify := closeNotifier.CloseNotify()

	go func() {
		<-closeNotify
		cancel(errors.New("client connection closed or lost"))
	}()

	return &Session{
		ctx:    ctx,
		logger: logger,
		cancel: cancel,
		events: make(chan []byte),
		writer: w,
	}, nil
}

func (s *Session) Send(data []byte) {
	if s.ctx.Err() != nil {
		return
	}

	select {
	case s.events <- data:
	case <-s.ctx.Done():
	}
}

func (s *Session) dispatch(tearDown func()) {
	defer tearDown()
	for {
		select {
		case data := <-s.events:
			message := fmt.Sprintf("data: %s\n\n", data)
			s.logger.Debug(strings.Trim(message, "\n"))
			_, err := fmt.Fprint(s.writer, message)
			if err != nil {
				s.logger.Debug("failed to write to session", "error", err)
				s.cancel(err)
				return
			}
			s.writer.(http.Flusher).Flush()
		case <-s.ctx.Done():
			return
		}
	}
}

// Wait waits until the client session has ended
func (s *Session) Wait() {
	<-s.ctx.Done()
}

type Server struct {
	ctx      context.Context
	logger   *slog.Logger
	sessions []*Session
	sync.RWMutex
}

func New(ctx context.Context, logger *slog.Logger) *Server {
	return &Server{
		ctx:      ctx,
		logger:   logger,
		sessions: make([]*Session, 0),
	}
}

func (s *Server) NewSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	s.Lock()
	defer s.Unlock()

	session, err := newSession(s.ctx, s.logger, w)
	if err != nil {
		return nil, err
	}
	s.sessions = append(s.sessions, session)

	go session.dispatch(func() {
		s.Lock()
		defer s.Unlock()

		for i := 0; i < len(s.sessions); i++ {
			if session == s.sessions[i] {
				s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
				s.logger.Info("session", "count", len(s.sessions))
				break
			}
		}
	})

	s.logger.Info("session", "count", len(s.sessions))

	return session, nil
}

func (s *Server) Broadcast(data []byte) {
	s.RLock()
	defer s.RUnlock()

	for _, session := range s.sessions {
		session.Send(data)
	}
}
