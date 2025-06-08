package customer

import (
	"bank/entity"
	"bank/internal/request"
	"bank/internal/response"
	"errors"
	"net/http"

	"github.com/shopspring/decimal"
)

type CreateTransferFundsRequest struct {
	SourceAccountID      uint64          `json:"source_account_id" validate:"required,number"`
	DestinationAccountID uint64          `json:"destination_account_id" validate:"required,number"`
	Amount               decimal.Decimal `json:"amount" validate:"required,decimal_required,decimal_positive,decimal_precision=6"`
}

func (h *Handler) CreateTransferFunds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateTransferFundsRequest
		if err := request.BindJSON(r, &req); err != nil {
			response.JsonError(w, http.StatusBadRequest, "invalid request")
			return
		}

		_, err := h.transactionDomain.CreateTransferFunds(r.Context(), entity.CreateTransferFundsParams{
			SourceAccountID:      req.SourceAccountID,
			DestinationAccountID: req.DestinationAccountID,
			Amount:               req.Amount,
		})
		if err != nil {
			switch {
			case errors.Is(err, entity.ErrInsufficientFunds):
				response.JsonError(w, http.StatusBadRequest, "your account has insufficient funds")
			case errors.Is(err, entity.ErrDataNotFound):
				response.JsonError(w, http.StatusBadRequest, "invalid account")
			case errors.Is(err, entity.ErrValidation):
				response.JsonError(w, http.StatusBadRequest, err.Error())
			default:
				response.JsonError(w, http.StatusInternalServerError, "it's not you, it's us. please contact support")
				h.logger.Error(r.Context(), "failed to create transfer funds: %v", err)
			}
			return
		}

		response.StatusOnly(w, http.StatusOK)
	}
}
