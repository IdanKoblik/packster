package auth

import (
	"artifactor/pkg/types"

	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) FetchToken(rawToken string) (*types.ApiToken, error) {
	args := m.Called(rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ApiToken), args.Error(1)
}

func (m *mockRepo) PruneToken(rawToken string) error {
	args := m.Called(rawToken)
	return args.Error(0)
}

func (m *mockRepo) CreateToken(request *types.RegisterRequest) (string, error) {
	args := m.Called(request)
	return args.String(0), args.Error(1)
}

func (m *mockRepo) TokenExists(rawToken string) (bool, error) {
	args := m.Called(rawToken)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepo) IsAdmin(rawToken string) (bool, error) {
	args := m.Called(rawToken)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepo) ListTokens() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
