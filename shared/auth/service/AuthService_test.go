package service

import (
	"errors"
	"testing"

	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/auth/domain/repositories"
)

type mockRepo struct {
	session *model.AuthSession
}

var _ repositories.SessionRepository = &mockRepo{}

func (m *mockRepo) Create(s *model.AuthSession) error { m.session = s; return nil }
func (m *mockRepo) FindBySessionID(id string) (*model.AuthSession, error) {
	if m.session != nil && m.session.SessionID == id {
		return m.session, nil
	}
	return nil, errors.New("not found")
}
func (m *mockRepo) Update(s *model.AuthSession) error                    { m.session = s; return nil }
func (m *mockRepo) DeleteBySessionID(id string) error                    { return nil }
func (m *mockRepo) Delete(s *model.AuthSession) error                    { return nil }
func (m *mockRepo) FindExpiredSessions() ([]*model.AuthSession, error)   { return nil, nil }
func (m *mockRepo) DeleteExpiredSessions() error                         { return nil }
func (m *mockRepo) FindByUserID(id string) ([]*model.AuthSession, error) { return nil, nil }
func (m *mockRepo) DeleteAllByUserID(id string) error                    { return nil }

func TestGenerateSessionID(t *testing.T) {
	svc := NewAuthService(&mockRepo{}, nil)
	id1, err := svc.generateSessionID()
	if err != nil || len(id1) == 0 {
		t.Fatalf("bad id %v %v", id1, err)
	}
	id2, _ := svc.generateSessionID()
	if id1 == id2 {
		t.Fatal("ids should differ")
	}
}

func TestCleanExpiredSessions(t *testing.T) {
	repo := &mockRepo{}
	svc := NewAuthService(repo, nil)
	if err := svc.CleanExpiredSessions(); err != nil {
		t.Fatal(err)
	}
}

func TestAccessors(t *testing.T) {
	repo := &mockRepo{}
	svc := NewAuthService(repo, nil)
	if svc.GetSessionRepo() != repo {
		t.Fatal("wrong repo")
	}
	if svc.GetAuthProvider() != nil {
		t.Fatal("expected nil auth provider")
	}
}
