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

func (m *mockProductRepo) DeleteProduct(name, group, token string, admin bool) error {
	args := m.Called(name, group, token, admin)
	return args.Error(0)
}

func (m *mockProductRepo) FetchProduct(name, group string) (*types.Product, error) {
	args := m.Called(name, group)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Product), args.Error(1)
}

func (m *mockProductRepo) DeleteToken(productName, group, sourceToken, targetToken string, admin bool) error {
	args := m.Called(productName, group, sourceToken, targetToken, admin)
	return args.Error(0)
}

func (m *mockProductRepo) AddToken(productName, group, sourceToken, targetToken string, permissions types.TokenPermissions, admin bool) error {
	args := m.Called(productName, group, sourceToken, targetToken, permissions, admin)
	return args.Error(0)
}

func (m *mockProductRepo) DeleteVersion(productName, group, version, token string, admin bool) error {
	args := m.Called(productName, group, version, token, admin)
	return args.Error(0)
}

func (m *mockProductRepo) AddVersion(productName, group, version, token string, admin bool, v types.Version) error {
	args := m.Called(productName, group, version, token, admin, v)
	return args.Error(0)
}

func (m *mockProductRepo) ListProducts() ([]types.Product, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Product), args.Error(1)
}

func (m *mockProductRepo) ListProductsByToken(hashedToken string) ([]types.Product, error) {
	args := m.Called(hashedToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Product), args.Error(1)
}
