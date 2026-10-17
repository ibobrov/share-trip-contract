-- +goose Up
CREATE TABLE contract_options (
    contract_id UUID NOT NULL,
    code TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    PRIMARY KEY (contract_id, code),
    CONSTRAINT contract_options_contract_id_fkey
        FOREIGN KEY (contract_id) REFERENCES contracts (id) ON DELETE CASCADE,
    CONSTRAINT contract_options_code_check
        CHECK (code IN ('trip_creation', 'trip_participants', 'notifications'))
);

-- +goose Down
DROP TABLE contract_options;
