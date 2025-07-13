package service

import (
	domain "github.com/r0x16/Raidark/shared/auth/domain"
	mod "github.com/r0x16/Raidark/shared/auth/domain/model"
	"testing"
)

// simple in-memory repo implementing repositories.SessionRepository
// minimal methods used here

type memRepo struct{ m map[string]*mod.AuthSession }

func newMemRepo() *memRepo                                             { return &memRepo{m: make(map[string]*mod.AuthSession)} }
func (r *memRepo) Create(s *mod.AuthSession) error                     { r.m[s.SessionID] = s; return nil }
func (r *memRepo) FindBySessionID(id string) (*mod.AuthSession, error) { return r.m[id], nil }
func (r *memRepo) Update(s *mod.AuthSession) error                     { r.m[s.SessionID] = s; return nil }
func (r *memRepo) DeleteBySessionID(id string) error                   { delete(r.m, id); return nil }
func (r *memRepo) Delete(s *mod.AuthSession) error                     { delete(r.m, s.SessionID); return nil }
func (r *memRepo) FindExpiredSessions() ([]*mod.AuthSession, error)    { return nil, nil }
func (r *memRepo) DeleteExpiredSessions() error                        { return nil }
func (r *memRepo) FindByUserID(id string) ([]*mod.AuthSession, error)  { return nil, nil }
func (r *memRepo) DeleteAllByUserID(id string) error                   { return nil }

// dummy auth provider

type dummyAuth struct{ domain.AuthProvider }

func (dummyAuth) Initialize() error { return nil }

func TestGenerateSessionID(t *testing.T) {
	repo := newMemRepo()
	s := NewAuthService(repo, dummyAuth{})
	id, err := s.generateSessionID()
	if err != nil || id == "" {
		t.Fatalf("invalid id")
	}
}
