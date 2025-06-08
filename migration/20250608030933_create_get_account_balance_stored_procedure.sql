-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION get_account_balance(
    filter_account_id BIGINT,
    filter_lock_for_update BOOLEAN DEFAULT FALSE
)
RETURNS DECIMAL(20, 6)
LANGUAGE plpgsql
AS $$
DECLARE
    v_balance DECIMAL(20, 6);
    v_last_transaction_id BIGINT;
    v_snapshot_balance DECIMAL(20, 6);
    v_transaction_delta DECIMAL(20, 6) := 0;
    rec RECORD;
BEGIN
    IF filter_lock_for_update THEN
        -- Lock the account record first
        PERFORM 1 FROM accounts WHERE id = filter_account_id FOR UPDATE;
        
        -- Get latest snapshot
        SELECT balance, last_transaction_id
        INTO v_snapshot_balance, v_last_transaction_id
        FROM account_balance_snapshots
        WHERE account_id = filter_account_id
        ORDER BY created_at DESC
        LIMIT 1;
        
        -- Lock and sum transactions individually
        -- Postgres does not support FOR UPDATE for aggregate queries
        FOR rec IN 
            SELECT amount, trx_type
            FROM transactions
            WHERE account_id = filter_account_id
              AND id > COALESCE(v_last_transaction_id, 0)
            ORDER BY id  -- Consistent ordering to avoid deadlocks
            FOR UPDATE
        LOOP
            CASE rec.trx_type
                WHEN 'CREDIT' THEN v_transaction_delta := v_transaction_delta + rec.amount;
                WHEN 'DEBIT' THEN v_transaction_delta := v_transaction_delta - rec.amount;
            END CASE;
        END LOOP;
        
    ELSE
        -- Non-locking version with aggregate
        SELECT balance, last_transaction_id
        INTO v_snapshot_balance, v_last_transaction_id
        FROM account_balance_snapshots
        WHERE account_id = filter_account_id
        ORDER BY created_at DESC
        LIMIT 1;
        
        SELECT COALESCE(
            SUM(CASE 
                WHEN trx_type = 'CREDIT' THEN amount
                WHEN trx_type = 'DEBIT' THEN -amount
                ELSE 0
            END), 0
        )
        INTO v_transaction_delta
        FROM transactions
        WHERE account_id = filter_account_id
          AND id > COALESCE(v_last_transaction_id, 0);
    END IF;
    
    v_balance := COALESCE(v_snapshot_balance, 0) + v_transaction_delta;
    RETURN v_balance;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_account_balance;
-- +goose StatementEnd
