-- +goose Up
CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,

    CONSTRAINT contracts_status_check
        CHECK (status IN ('draft', 'active', 'suspended', 'terminated')),
    CONSTRAINT contracts_period_check
        CHECK (expires_at IS NULL OR expires_at > starts_at)
);

CREATE INDEX contracts_company_id_idx ON contracts (company_id);

CREATE UNIQUE INDEX contracts_one_active_per_company_idx
    ON contracts (company_id)
    WHERE status = 'active';

-- +goose Down
DROP TABLE contracts;
