package endpoints

import (
	pkghttp "artifactor/pkg/http"
	"artifactor/pkg/tokens"

	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) FetchToken(rawToken string) (*tokens.ApiToken, error) {
	args := m.Called(rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tokens.ApiToken), args.Error(1)
}

func (m *mockRepo) PruneToken(rawToken string) error {
	args := m.Called(rawToken)
	return args.Error(0)
}

func (m *mockRepo) CreateToken(request *pkghttp.RegisterRequest) (string, error) {
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
