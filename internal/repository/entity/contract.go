package entity

import (
	"time"

	"github.com/google/uuid"
)

type Contract struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Status    string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []ContractOption
}

type ContractOption struct {
	Code    string
	Enabled bool
}
