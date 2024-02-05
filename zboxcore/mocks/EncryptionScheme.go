// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	encryption "github.com/0chain/gosdk/zboxcore/encryption"
	mock "github.com/stretchr/testify/mock"
)

// EncryptionScheme is an autogenerated mock type for the EncryptionScheme type
type EncryptionScheme struct {
	mock.Mock
}

// Decrypt provides a mock function with given fields: _a0
func (_m *EncryptionScheme) Decrypt(_a0 *encryption.EncryptedMessage) ([]byte, error) {
	ret := _m.Called(_a0)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(*encryption.EncryptedMessage) ([]byte, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*encryption.EncryptedMessage) []byte); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(*encryption.EncryptedMessage) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Encrypt provides a mock function with given fields: data
func (_m *EncryptionScheme) Encrypt(data []byte) (*encryption.EncryptedMessage, error) {
	ret := _m.Called(data)

	var r0 *encryption.EncryptedMessage
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte) (*encryption.EncryptedMessage, error)); ok {
		return rf(data)
	}
	if rf, ok := ret.Get(0).(func([]byte) *encryption.EncryptedMessage); ok {
		r0 = rf(data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*encryption.EncryptedMessage)
		}
	}

	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetEncryptedKey provides a mock function with given fields:
func (_m *EncryptionScheme) GetEncryptedKey() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetEncryptedKeyPoint provides a mock function with given fields:
func (_m *EncryptionScheme) GetEncryptedKeyPoint() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPrivateKey provides a mock function with given fields:
func (_m *EncryptionScheme) GetPrivateKey() (string, error) {
	ret := _m.Called()

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPublicKey provides a mock function with given fields:
func (_m *EncryptionScheme) GetPublicKey() (string, error) {
	ret := _m.Called()

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func() (string, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReGenKey provides a mock function with given fields: encPublicKey, tag
func (_m *EncryptionScheme) GetReGenKey(encPublicKey string, tag string) (string, error) {
	ret := _m.Called(encPublicKey, tag)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(encPublicKey, tag)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(encPublicKey, tag)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(encPublicKey, tag)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InitForDecryption provides a mock function with given fields: tag, encryptedKey
func (_m *EncryptionScheme) InitForDecryption(tag string, encryptedKey string) error {
	ret := _m.Called(tag, encryptedKey)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(tag, encryptedKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InitForDecryptionWithPoint provides a mock function with given fields: tag, point
func (_m *EncryptionScheme) InitForDecryptionWithPoint(tag string, point string) error {
	ret := _m.Called(tag, point)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(tag, point)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InitForEncryption provides a mock function with given fields: tag
func (_m *EncryptionScheme) InitForEncryption(tag string) {
	_m.Called(tag)
}

// InitForEncryptionWithPoint provides a mock function with given fields: tag, point
func (_m *EncryptionScheme) InitForEncryptionWithPoint(tag string, point string) error {
	ret := _m.Called(tag, point)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(tag, point)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Initialize provides a mock function with given fields: mnemonic
func (_m *EncryptionScheme) Initialize(mnemonic string) ([]byte, error) {
	ret := _m.Called(mnemonic)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]byte, error)); ok {
		return rf(mnemonic)
	}
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(mnemonic)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(mnemonic)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InitializeWithPrivateKey provides a mock function with given fields: privateKey
func (_m *EncryptionScheme) InitializeWithPrivateKey(privateKey []byte) error {
	ret := _m.Called(privateKey)

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(privateKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReDecrypt provides a mock function with given fields: D
func (_m *EncryptionScheme) ReDecrypt(D *encryption.ReEncryptedMessage) ([]byte, error) {
	ret := _m.Called(D)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(*encryption.ReEncryptedMessage) ([]byte, error)); ok {
		return rf(D)
	}
	if rf, ok := ret.Get(0).(func(*encryption.ReEncryptedMessage) []byte); ok {
		r0 = rf(D)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(*encryption.ReEncryptedMessage) error); ok {
		r1 = rf(D)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReEncrypt provides a mock function with given fields: encMsg, reGenKey, clientPublicKey
func (_m *EncryptionScheme) ReEncrypt(encMsg *encryption.EncryptedMessage, reGenKey string, clientPublicKey string) (*encryption.ReEncryptedMessage, error) {
	ret := _m.Called(encMsg, reGenKey, clientPublicKey)

	var r0 *encryption.ReEncryptedMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(*encryption.EncryptedMessage, string, string) (*encryption.ReEncryptedMessage, error)); ok {
		return rf(encMsg, reGenKey, clientPublicKey)
	}
	if rf, ok := ret.Get(0).(func(*encryption.EncryptedMessage, string, string) *encryption.ReEncryptedMessage); ok {
		r0 = rf(encMsg, reGenKey, clientPublicKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*encryption.ReEncryptedMessage)
		}
	}

	if rf, ok := ret.Get(1).(func(*encryption.EncryptedMessage, string, string) error); ok {
		r1 = rf(encMsg, reGenKey, clientPublicKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewEncryptionScheme creates a new instance of EncryptionScheme. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEncryptionScheme(t interface {
	mock.TestingT
	Cleanup(func())
}) *EncryptionScheme {
	mock := &EncryptionScheme{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
