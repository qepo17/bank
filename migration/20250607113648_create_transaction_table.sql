-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS transactions (
    id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    account_id bigint NOT NULL,
    transfer_id bigint, -- for nested transactions
    amount decimal(20, 6) NOT NULL,
    trx_type varchar NOT NULL, -- enum: CREDIT, DEBIT
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id),
    FOREIGN KEY (transfer_id) REFERENCES transfers(id)
);
CREATE INDEX IF NOT EXISTS idx_transactions_account_id_created_at ON transactions (account_id, created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_transfer_id ON transactions (transfer_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd
