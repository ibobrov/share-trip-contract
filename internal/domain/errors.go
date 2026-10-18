package domain

import "errors"

var (
	ErrBusinessValidation   = errors.New("нарушены условия договора")
	ErrNotFound             = errors.New("договор не найден")
	ErrNoActiveContract     = errors.New("действующий активный договор не найден")
	ErrInvalidStatus        = errors.New("активировать можно только договор в статусе draft")
	ErrActiveContractExists = errors.New("у компании уже есть активный договор")
)

type BusinessValidationError struct {
	Message string
}

func (e *BusinessValidationError) Error() string { return e.Message }
func (e *BusinessValidationError) Unwrap() error { return ErrBusinessValidation }
