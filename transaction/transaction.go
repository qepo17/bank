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

const (
	// PostgreSQL composite type parsing constants
	compositeTypeFields   = 3
	postgresqlBooleanTrue = "t"
	postgresqlNullValue   = "<NULL>"
	quoteMark             = "\""
	escapedQuote          = "\\\""
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
		return entity.CreateTransferFundsResult{}, d.mapTransferError(transferFundsResult.ErrorMessage)
	}

	return transferFundsResult, nil
}

// parseTransferFundsResult parses a PostgreSQL composite type string into a CreateTransferFundsResult
// The input format is: "(transfer_id,success,error_message)"
// Example: "(16007,t,\"Transfer completed successfully\")"
func (d *TransactionDomain) parseTransferFundsResult(result interface{}) (entity.CreateTransferFundsResult, error) {
	resultStr, err := d.convertToString(result)
	if err != nil {
		return entity.CreateTransferFundsResult{}, err
	}

	content, err := d.extractCompositeContent(resultStr)
	if err != nil {
		return entity.CreateTransferFundsResult{}, err
	}

	fields := strings.Split(content, ",")
	if len(fields) != compositeTypeFields {
		return entity.CreateTransferFundsResult{}, fmt.Errorf("invalid composite type content, expected %d fields got %d: %s", compositeTypeFields, len(fields), content)
	}

	transferID, err := d.parseTransferID(fields[0])
	if err != nil {
		return entity.CreateTransferFundsResult{}, err
	}

	success := d.parseSuccess(fields[1])
	errorMessage := d.parseErrorMessage(fields[2])

	return entity.CreateTransferFundsResult{
		TransferID:   transferID,
		Success:      success,
		ErrorMessage: errorMessage,
	}, nil
}

// convertToString converts the result to string, handling both string and []byte cases
func (d *TransactionDomain) convertToString(result interface{}) (string, error) {
	switch v := result.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("expected string or []byte result, got %T", result)
	}
}

// extractCompositeContent removes parentheses and returns the inner content
func (d *TransactionDomain) extractCompositeContent(resultStr string) (string, error) {
	if !strings.HasPrefix(resultStr, "(") || !strings.HasSuffix(resultStr, ")") {
		return "", fmt.Errorf("invalid composite type format, expected parentheses: %s", resultStr)
	}
	return resultStr[1 : len(resultStr)-1], nil
}

// parseTransferID parses the transfer_id field
func (d *TransactionDomain) parseTransferID(field string) (uint64, error) {
	transferIDStr := strings.TrimSpace(field)
	if transferIDStr == "" || transferIDStr == postgresqlNullValue {
		return 0, nil
	}

	parsed, err := strconv.ParseUint(transferIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid transfer_id: %s", transferIDStr)
	}
	return parsed, nil
}

// parseSuccess parses the success field
func (d *TransactionDomain) parseSuccess(field string) bool {
	successStr := strings.TrimSpace(field)
	return successStr == postgresqlBooleanTrue
}

// parseErrorMessage parses the error_message field, removing quotes and unescaping
func (d *TransactionDomain) parseErrorMessage(field string) string {
	errorMessage := strings.TrimSpace(field)
	if strings.HasPrefix(errorMessage, quoteMark) && strings.HasSuffix(errorMessage, quoteMark) {
		errorMessage = errorMessage[1 : len(errorMessage)-1]
		// Unescape any escaped quotes
		errorMessage = strings.ReplaceAll(errorMessage, escapedQuote, quoteMark)
	}
	return errorMessage
}

// mapTransferError maps database error messages to appropriate domain errors
func (d *TransactionDomain) mapTransferError(errorMessage string) error {
	normalizedErr := strings.ToLower(errorMessage)
	switch {
	case strings.Contains(normalizedErr, "insufficient funds"):
		return entity.ErrInsufficientFunds
	case strings.Contains(normalizedErr, "account does not exist"):
		return entity.ErrDataNotFound
	case strings.Contains(normalizedErr, "transfer amount must be positive"),
		strings.Contains(normalizedErr, "cannot transfer to the same account"):
		return fmt.Errorf("%w: %s", entity.ErrValidation, errorMessage)
	default:
		return fmt.Errorf("transfer funds failed: %s", errorMessage)
	}
}
