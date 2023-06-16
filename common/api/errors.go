package api

import (
	"fmt"
)

// region - ErrClientNotFound

type ErrClientNotFound struct {
	message string
}

func NewClientNotFoundError(clientId string) error {
	return &ErrClientNotFound{
		message: fmt.Sprintf("client %s not found", clientId),
	}
}
func (e *ErrClientNotFound) Error() string {
	return e.message
}
func (e *ErrClientNotFound) Is(tgt error) bool {
	_, ok := tgt.(*ErrClientNotFound)
	if !ok {
		return false
	}
	return true
}

// endregion
// region - ErrTenantNotFound

type ErrTenantNotFound struct {
	message string
}

func NewTenantNotFoundError(tenant string) error {
	return &ErrTenantNotFound{
		message: fmt.Sprintf("tenant %s not found", tenant),
	}
}
func (e *ErrTenantNotFound) Error() string {
	return e.message
}
func (e *ErrTenantNotFound) Is(tgt error) bool {
	_, ok := tgt.(*ErrTenantNotFound)
	if !ok {
		return false
	}
	return true
}

// endregion
// region - ErrTenantsClientNotFound

type ErrTenantsClientNotFound struct {
	message string
}

func NewTenantsClientNotFoundError(clientId string) error {
	return &ErrTenantsClientNotFound{
		message: fmt.Sprintf("tenant's client %s not found", clientId),
	}
}
func (e *ErrTenantsClientNotFound) Error() string {
	return e.message
}
func (e *ErrTenantsClientNotFound) Is(tgt error) bool {
	_, ok := tgt.(*ErrTenantsClientNotFound)
	if !ok {
		return false
	}
	return true
}

// endregion
// region - ErrAlreadyRegistered

type ErrAlreadyRegistered struct {
	message string
}

func NewAlreadyRegisteredError() error {
	return &ErrTenantsClientNotFound{
		message: fmt.Sprintf("client already registered"),
	}
}
func (e *ErrAlreadyRegistered) Error() string {
	return e.message
}
func (e *ErrAlreadyRegistered) Is(tgt error) bool {
	_, ok := tgt.(*ErrAlreadyRegistered)
	if !ok {
		return false
	}
	return true
}

// endregion
// region - ErrMaxClientsReached

type ErrMaxClientsReached struct {
	message string
}

func NewMaxClientsReachedError(max int) error {
	return &ErrMaxClientsReached{
		message: fmt.Sprintf("maximum clients reached (%d)", max),
	}
}
func (e *ErrMaxClientsReached) Error() string {
	return e.message
}
func (e *ErrMaxClientsReached) Is(tgt error) bool {
	_, ok := tgt.(*ErrClientNotFound)
	if !ok {
		return false
	}
	return true
}

// endregion
