-- +goose Up
-- +goose StatementBegin
ALTER TABLE account_balance_snapshots
ADD CONSTRAINT fk_account_balance_snapshots_last_transaction_id
FOREIGN KEY (last_transaction_id) REFERENCES transactions(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE account_balance_snapshots
DROP CONSTRAINT fk_account_balance_snapshots_last_transaction_id;
-- +goose StatementEnd
