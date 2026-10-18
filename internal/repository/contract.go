package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/repository/entity"
	"github.com/jackc/pgx/v5"
)

type ContractRepository struct{}

func NewContractRepository() *ContractRepository { return &ContractRepository{} }

func (r *ContractRepository) Create(ctx context.Context, tx pgx.Tx, contract entity.Contract) error {
	_, err := tx.Exec(ctx, `INSERT INTO contracts (id, company_id, status, starts_at, expires_at)
        VALUES ($1, $2, $3, $4, $5)`, contract.ID, contract.CompanyID, contract.Status, contract.StartsAt, contract.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert contract: %w", err)
	}
	for _, option := range contract.Options {
		if _, err := tx.Exec(ctx, `INSERT INTO contract_options (contract_id, code, enabled) VALUES ($1, $2, $3)`, contract.ID, option.Code, option.Enabled); err != nil {
			return fmt.Errorf("insert contract option: %w", err)
		}
	}
	return nil
}

func (r *ContractRepository) Get(ctx context.Context, tx pgx.Tx, id uuid.UUID, lock bool) (entity.Contract, error) {
	query := `SELECT id, company_id, status, starts_at, expires_at FROM contracts WHERE id = $1`
	if lock {
		query += ` FOR UPDATE`
	}
	return r.scan(ctx, tx, tx.QueryRow(ctx, query, id))
}

// Active returns the contract by status, including expired and future contracts.
// The domain decides whether its dates permit the requested operation.
func (r *ContractRepository) Active(ctx context.Context, tx pgx.Tx, companyID uuid.UUID) (entity.Contract, error) {
	return r.scan(ctx, tx, tx.QueryRow(ctx, `SELECT id, company_id, status, starts_at, expires_at
        FROM contracts WHERE company_id = $1 AND status = 'active'`, companyID))
}

func (r *ContractRepository) CompanyHasContracts(ctx context.Context, tx pgx.Tx, companyID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT FROM contracts WHERE company_id = $1)`, companyID).Scan(&exists)
	return exists, err
}

func (r *ContractRepository) Activate(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	tag, err := tx.Exec(ctx, `UPDATE contracts SET status = 'active' WHERE id = $1 AND status = 'draft'`, id)
	if err != nil {
		return fmt.Errorf("activate contract: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("activate contract: %w", pgx.ErrNoRows)
	}
	return nil
}

func (r *ContractRepository) scan(ctx context.Context, tx pgx.Tx, row pgx.Row) (entity.Contract, error) {
	var contract entity.Contract
	if err := row.Scan(&contract.ID, &contract.CompanyID, &contract.Status, &contract.StartsAt, &contract.ExpiresAt); err != nil {
		return entity.Contract{}, fmt.Errorf("read contract: %w", err)
	}
	contract.StartsAt = contract.StartsAt.UTC()
	if contract.ExpiresAt != nil {
		t := contract.ExpiresAt.UTC()
		contract.ExpiresAt = &t
	}
	rows, err := tx.Query(ctx, `SELECT code, enabled FROM contract_options WHERE contract_id = $1 ORDER BY code`, contract.ID)
	if err != nil {
		return entity.Contract{}, fmt.Errorf("read contract options: %w", err)
	}
	defer rows.Close()
	contract.Options = make([]entity.ContractOption, 0)
	for rows.Next() {
		var option entity.ContractOption
		if err := rows.Scan(&option.Code, &option.Enabled); err != nil {
			return entity.Contract{}, fmt.Errorf("scan contract option: %w", err)
		}
		contract.Options = append(contract.Options, option)
	}
	if err := rows.Err(); err != nil {
		return entity.Contract{}, fmt.Errorf("iterate contract options: %w", err)
	}
	return contract, nil
}
