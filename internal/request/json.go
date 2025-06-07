package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var validate = validator.New()

const maxRequestSize = 1 << 20 // 1MB

func init() {
	// Register custom validators for decimal.Decimal
	validate.RegisterValidation("decimal_required", validateDecimalRequired)
	validate.RegisterValidation("decimal_positive", validateDecimalPositive)
	validate.RegisterValidation("decimal_non_negative", validateDecimalNonNegative)
	validate.RegisterValidation("decimal_min", validateDecimalMin)
	validate.RegisterValidation("decimal_max", validateDecimalMax)
	validate.RegisterValidation("decimal_precision", validateDecimalPrecision)
}

func BindJSON[T any](r *http.Request, v *T) error {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("content-type must be application/json")
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestSize)

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := validate.Struct(v); err != nil {
		return errors.New(formatValidationError(err))
	}

	return nil
}

// formatValidationError converts validator errors to human-readable messages
func formatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			return getHumanReadableError(fieldError)
		}
	}
	return "validation failed"
}

// getHumanReadableError converts a single field error to human-readable message
func getHumanReadableError(fe validator.FieldError) string {
	fieldName := getFieldDisplayName(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "number":
		return fmt.Sprintf("%s must be a valid number", fieldName)
	case "decimal_required":
		return fmt.Sprintf("%s is required", fieldName)
	case "decimal_positive":
		return fmt.Sprintf("%s must be greater than 0", fieldName)
	case "decimal_non_negative":
		return fmt.Sprintf("%s must be greater than or equal to 0", fieldName)
	case "decimal_min":
		return fmt.Sprintf("%s must be at least %s", fieldName, fe.Param())
	case "decimal_max":
		return fmt.Sprintf("%s must be at most %s", fieldName, fe.Param())
	case "decimal_precision":
		return fmt.Sprintf("%s has too many decimal places (max %s)", fieldName, fe.Param())
	case "min":
		return fmt.Sprintf("%s must be at least %s", fieldName, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fieldName, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fieldName, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldName, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fieldName, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fieldName, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fieldName)
	}
}

// getFieldDisplayName converts struct field names to human-readable names
func getFieldDisplayName(fieldName string) string {
	switch fieldName {
	case "AccountID":
		return "account id"
	case "InitialBalance":
		return "initial balance"
	case "Balance":
		return "balance"
	case "Amount":
		return "amount"
	default:
		// Convert PascalCase to space-separated lowercase
		result := ""
		for i, r := range fieldName {
			if i > 0 && r >= 'A' && r <= 'Z' {
				result += " "
			}
			result += strings.ToLower(string(r))
		}
		return result
	}
}

// validateDecimalRequired checks if decimal is not zero (equivalent to required)
func validateDecimalRequired(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return !dec.IsZero()
	}
	return false
}

// validateDecimalPositive checks if decimal is greater than zero
func validateDecimalPositive(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return dec.IsPositive()
	}
	return false
}

// validateDecimalNonNegative checks if decimal is greater than or equal to zero
func validateDecimalNonNegative(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		return !dec.IsNegative()
	}
	return false
}

// validateDecimalMin checks if decimal is greater than or equal to minimum value
func validateDecimalMin(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		minStr := fl.Param()
		if min, err := decimal.NewFromString(minStr); err == nil {
			return dec.GreaterThanOrEqual(min)
		}
	}
	return false
}

// validateDecimalMax checks if decimal is less than or equal to maximum value
func validateDecimalMax(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		maxStr := fl.Param()
		if max, err := decimal.NewFromString(maxStr); err == nil {
			return dec.LessThanOrEqual(max)
		}
	}
	return false
}

// validateDecimalPrecision checks if decimal has at most specified decimal places
func validateDecimalPrecision(fl validator.FieldLevel) bool {
	if dec, ok := fl.Field().Interface().(decimal.Decimal); ok {
		precisionStr := fl.Param()
		if precision, err := strconv.Atoi(precisionStr); err == nil {
			// Get the number of decimal places
			decimalPlaces := dec.Exponent() * -1
			if decimalPlaces < 0 {
				decimalPlaces = 0
			}
			return int(decimalPlaces) <= precision
		}
	}
	return false
}
