package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockDoType func(req *http.Request) (*http.Response, error)

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
