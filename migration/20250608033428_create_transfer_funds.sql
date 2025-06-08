-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION transfer_funds(
    param_from_account_id BIGINT,
    param_to_account_id BIGINT,
    param_amount DECIMAL(20,6)
)
RETURNS TABLE(
    transfer_id BIGINT,
    success BOOLEAN,
    error_message TEXT
) 
LANGUAGE plpgsql
AS $$
DECLARE
    v_transfer_id BIGINT;
    v_from_balance DECIMAL(20,6);
    v_first_account BIGINT;
    v_second_account BIGINT;
BEGIN
    -- Validate input parameters
    IF param_amount <= 0 THEN
        RETURN QUERY SELECT NULL::BIGINT, FALSE, 'Transfer amount must be positive';
        RETURN;
    END IF;
    
    IF param_from_account_id = param_to_account_id THEN
        RETURN QUERY SELECT NULL::BIGINT, FALSE, 'Cannot transfer to the same account';
        RETURN;
    END IF;
    
    -- Lock accounts in consistent order to prevent deadlocks
    v_first_account := LEAST(param_from_account_id, param_to_account_id);
    v_second_account := GREATEST(param_from_account_id, param_to_account_id);
    
    -- Lock both accounts in order
    PERFORM 1 FROM accounts WHERE id = v_first_account FOR UPDATE;
    PERFORM 1 FROM accounts WHERE id = v_second_account FOR UPDATE;
    
    -- Verify both accounts exist
    IF NOT EXISTS (SELECT 1 FROM accounts WHERE id = param_from_account_id) THEN
        RETURN QUERY SELECT NULL::BIGINT, FALSE, 'From account does not exist';
        RETURN;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM accounts WHERE id = param_to_account_id) THEN
        RETURN QUERY SELECT NULL::BIGINT, FALSE, 'To account does not exist';
        RETURN;
    END IF;

    -- Get and lock account's balance
    SELECT get_account_balance(param_from_account_id, true) INTO v_from_balance;
    
    -- Check sufficient funds
    IF v_from_balance IS NULL OR v_from_balance < param_amount THEN
        RETURN QUERY SELECT NULL::BIGINT, FALSE, 'Insufficient funds';
        RETURN;
    END IF;
    
    -- Create transfer record
    INSERT INTO transfers (from_account_id, to_account_id)
    VALUES (param_from_account_id, param_to_account_id)
    RETURNING id INTO v_transfer_id;
    
    -- Create transactions atomically
    INSERT INTO transactions (account_id, transfer_id, amount, trx_type)
    VALUES 
        (param_from_account_id, v_transfer_id, param_amount, 'DEBIT'),
        (param_to_account_id, v_transfer_id, param_amount, 'CREDIT');
    
    RETURN QUERY SELECT v_transfer_id, TRUE, 'Transfer completed successfully'::TEXT;
    
EXCEPTION
    WHEN OTHERS THEN
        RETURN QUERY SELECT NULL::BIGINT, FALSE, SQLERRM;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS transfer_funds;
-- +goose StatementEnd
