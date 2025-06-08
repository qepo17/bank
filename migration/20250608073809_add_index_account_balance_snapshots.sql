-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_account_balance_snapshots_account_id_created_at ON account_balance_snapshots (account_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_account_balance_snapshots_account_id_created_at;
-- +goose StatementEnd
