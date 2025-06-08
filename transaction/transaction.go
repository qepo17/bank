package transaction

import (
	"bank/entity"
	"bank/internal/db/sqlc"
	"bank/internal/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type TransactionDomain struct {
	db      *sql.DB
	queries *sqlc.Queries
	logger  *logger.Logger
}

func NewTransactionDomain(db *sql.DB, sqlc *sqlc.Queries, logger *logger.Logger) (*TransactionDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if sqlc == nil {
		return nil, errors.New("sqlc is nil")
	}

	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	log := logger.WithField("domain", "transaction")
	return &TransactionDomain{db: db, queries: sqlc, logger: log}, nil
}

func (d *TransactionDomain) CreateTransferFunds(ctx context.Context, param entity.CreateTransferFundsParams) (entity.CreateTransferFundsResult, error) {
	transferFunds, err := d.queries.CreateTransferTransaction(ctx, sqlc.CreateTransferTransactionParams{
		ParamFromAccountID: int64(param.SourceAccountID),
		ParamToAccountID:   int64(param.DestinationAccountID),
		ParamAmount:        param.Amount.String(),
	})
	if err != nil {
		d.logger.Error(ctx, "param=%+v, error=%v", param, err)
		return entity.CreateTransferFundsResult{}, fmt.Errorf("failed to create transfer funds: %w", err)
	}

	if transferFunds == nil {
		d.logger.Error(ctx, "param=%+v, invalid transfer funds result", param)
		return entity.CreateTransferFundsResult{}, fmt.Errorf("invalid transfer funds result")
	}

	transferFundsResult, err := d.parseTransferFundsResult(transferFunds)
	if err != nil {
		d.logger.Error(ctx, "param=%+v, failed to parse result: %v", param, err)
		return entity.CreateTransferFundsResult{}, fmt.Errorf("failed to parse transfer funds result: %w", err)
	}

	if !transferFundsResult.Success {
		d.logger.Error(ctx, "param=%+v, transfer funds failed", param)
		normalizedErr := strings.ToLower(transferFundsResult.ErrorMessage)
		switch {
		case strings.Contains(normalizedErr, "insufficient funds"):
			return entity.CreateTransferFundsResult{}, entity.ErrInsufficientFunds
		case strings.Contains(normalizedErr, "account does not exist"):
			return entity.CreateTransferFundsResult{}, entity.ErrDataNotFound
		case strings.Contains(normalizedErr, "transfer amount must be positive"),
			strings.Contains(normalizedErr, "cannot transfer to the same account"):
			return entity.CreateTransferFundsResult{}, fmt.Errorf("%w: %s", entity.ErrValidation, transferFundsResult.ErrorMessage)
		default:
			return entity.CreateTransferFundsResult{}, fmt.Errorf("transfer funds failed: %s", transferFundsResult.ErrorMessage)
		}
	}

	return transferFundsResult, nil
}

// parseTransferFundsResult parses a PostgreSQL composite type string into a CreateTransferFundsResult
// The input format is: "(transfer_id,success,error_message)"
// Example: "(16007,t,\"Transfer completed successfully\")"
func (d *TransactionDomain) parseTransferFundsResult(result interface{}) (entity.CreateTransferFundsResult, error) {
	// Convert result to string (handle both string and []byte cases)
	var resultStr string
	switch v := result.(type) {
	case string:
		resultStr = v
	case []byte:
		resultStr = string(v)
	default:
		return entity.CreateTransferFundsResult{}, fmt.Errorf("expected string or []byte result, got %T", result)
	}

	// Remove parentheses
	if !strings.HasPrefix(resultStr, "(") || !strings.HasSuffix(resultStr, ")") {
		return entity.CreateTransferFundsResult{}, fmt.Errorf("invalid composite type format: %s", resultStr)
	}
	content := resultStr[1 : len(resultStr)-1]

	results := strings.Split(content, ",")
	if len(results) != 3 {
		return entity.CreateTransferFundsResult{}, fmt.Errorf("invalid composite type content: %s", content)
	}

	// Parse transfer_id
	transferIDStr := strings.TrimSpace(results[0])
	var transferID uint64
	if transferIDStr != "" && transferIDStr != "<NULL>" {
		parsed, err := strconv.ParseUint(transferIDStr, 10, 64)
		if err != nil {
			return entity.CreateTransferFundsResult{}, fmt.Errorf("invalid transfer_id: %s", transferIDStr)
		}
		transferID = parsed
	}

	// Parse success
	successStr := strings.TrimSpace(results[1])
	success := successStr == "t"

	// Parse error_message (remove quotes if present)
	errorMessage := strings.TrimSpace(results[2])
	if strings.HasPrefix(errorMessage, "\"") && strings.HasSuffix(errorMessage, "\"") {
		errorMessage = errorMessage[1 : len(errorMessage)-1]
		// Unescape any escaped quotes
		errorMessage = strings.ReplaceAll(errorMessage, "\\\"", "\"")
	}

	return entity.CreateTransferFundsResult{
		TransferID:   transferID,
		Success:      success,
		ErrorMessage: errorMessage,
	}, nil
}
