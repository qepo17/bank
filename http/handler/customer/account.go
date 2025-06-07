package customer

import (
	"bank/entity"
	"bank/internal/request"
	"bank/internal/response"
	"errors"
	"net/http"

	"github.com/shopspring/decimal"
)

type CreateAccountRequest struct {
	AccountID      uint64          `json:"account_id" validate:"required,number"`
	InitialBalance decimal.Decimal `json:"initial_balance" validate:"decimal_required,decimal_positive,decimal_precision=6"`
}

func (h *Handler) CreateAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateAccountRequest
		if err := request.BindJSON(r, &req); err != nil {
			response.JsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		account := &entity.CreateAccount{
			AccountID:      req.AccountID,
			InitialBalance: req.InitialBalance,
		}

		if err := h.accountDomain.CreateAccount(r.Context(), account); err != nil {
			switch {
			case errors.Is(err, entity.ErrValidation):
				response.JsonError(w, http.StatusBadRequest, err.Error())
			default:
				response.JsonError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		response.StatusOnly(w, http.StatusCreated)
	}
}
