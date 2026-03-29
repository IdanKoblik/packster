package product

import (
	"packster/pkg/types"

	"github.com/stretchr/testify/mock"
)

type mockProductRepo struct {
	mock.Mock
}

func (m *mockProductRepo) CreateProduct(product *types.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockProductRepo) DeleteProduct(name, token string, admin bool) error {
	args := m.Called(name, token, admin)
	return args.Error(0)
}

func (m *mockProductRepo) FetchProduct(name string) (*types.Product, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Product), args.Error(1)
}

func (m *mockProductRepo) DeleteToken(productName, sourceToken, targetToken string, admin bool) error {
	args := m.Called(productName, sourceToken, targetToken, admin)
	return args.Error(0)
}

func (m *mockProductRepo) AddToken(productName, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error {
	args := m.Called(productName, sourceToken, targetToken, permissions, admin)
	return args.Error(0)
}

func (m *mockProductRepo) DeleteVersion(productName, version, token string, admin bool) error {
	args := m.Called(productName, version, token, admin)
	return args.Error(0)
}

func (m *mockProductRepo) AddVersion(productName, version, token string, admin bool, v types.Version) error {
	args := m.Called(productName, version, token, admin, v)
	return args.Error(0)
}

func (m *mockProductRepo) ListProducts() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockProductRepo) ListProductsByToken(hashedToken string) ([]string, error) {
	args := m.Called(hashedToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
