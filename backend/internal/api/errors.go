package api

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"minibankingplatform/internal/domain"
)

const problemBaseURL = "https://minibankingplatform.com/problems/"

// MapError converts a domain error to ProblemDetails with appropriate HTTP status code.
func MapError(err error, instance string) (ProblemDetails, int) {
	problem := ProblemDetails{
		Instance: ptr(instance),
	}

	// Validation errors
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		problem.Type = problemBaseURL + "validation-error"
		problem.Title = "Validation Error"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr(validationErrs.Error())
		return problem, http.StatusBadRequest
	}

	// User already exists
	var userExistsErr *domain.UserAlreadyExistsError
	if errors.As(err, &userExistsErr) {
		problem.Type = problemBaseURL + "user-already-exists"
		problem.Title = "User Already Exists"
		problem.Status = http.StatusConflict
		problem.Detail = ptr(userExistsErr.Error())
		problem.Set("email", userExistsErr.Email)
		return problem, http.StatusConflict
	}

	// Invalid credentials
	var invalidCredsErr *domain.InvalidCredentialsError
	if errors.As(err, &invalidCredsErr) {
		problem.Type = problemBaseURL + "invalid-credentials"
		problem.Title = "Invalid Credentials"
		problem.Status = http.StatusUnauthorized
		problem.Detail = ptr("The provided email or password is incorrect")
		return problem, http.StatusUnauthorized
	}

	// Account not found
	var accountNotFoundErr *domain.AccountNotFoundError
	if errors.As(err, &accountNotFoundErr) {
		problem.Type = problemBaseURL + "account-not-found"
		problem.Title = "Account Not Found"
		problem.Status = http.StatusNotFound
		problem.Detail = ptr(accountNotFoundErr.Error())
		problem.Set("accountId", uuid.UUID(accountNotFoundErr.AccountID).String())
		return problem, http.StatusNotFound
	}

	// Insufficient funds
	var insufficientFundsErr *domain.InsufficientFundsError
	if errors.As(err, &insufficientFundsErr) {
		problem.Type = problemBaseURL + "insufficient-funds"
		problem.Title = "Insufficient Funds"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr("Account has insufficient funds for this operation")
		problem.Set("available", insufficientFundsErr.AvailableBalance.String())
		problem.Set("required", insufficientFundsErr.RequestedAmount.String())
		return problem, http.StatusBadRequest
	}

	// Currency mismatch
	var currencyMismatchErr *domain.CurrencyMismatchError
	if errors.As(err, &currencyMismatchErr) {
		problem.Type = problemBaseURL + "currency-mismatch"
		problem.Title = "Currency Mismatch"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr(currencyMismatchErr.Error())
		return problem, http.StatusBadRequest
	}

	// Negative transfer amount
	var negativeTransferErr *domain.NegativeTransferError
	if errors.As(err, &negativeTransferErr) {
		problem.Type = problemBaseURL + "negative-transfer-amount"
		problem.Title = "Negative Transfer Amount"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr(negativeTransferErr.Error())
		return problem, http.StatusBadRequest
	}

	// Negative exchange amount
	var negativeExchangeErr *domain.NegativeExchangeError
	if errors.As(err, &negativeExchangeErr) {
		problem.Type = problemBaseURL + "negative-exchange-amount"
		problem.Title = "Negative Exchange Amount"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr(negativeExchangeErr.Error())
		return problem, http.StatusBadRequest
	}

	// Same currency exchange
	var sameCurrencyExchangeErr *domain.SameCurrencyExchangeError
	if errors.As(err, &sameCurrencyExchangeErr) {
		problem.Type = problemBaseURL + "same-currency-exchange"
		problem.Title = "Same Currency Exchange"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr(sameCurrencyExchangeErr.Error())
		return problem, http.StatusBadRequest
	}

	// Same currency exchange rate
	var sameCurrencyExchangeRateErr *domain.SameCurrencyExchangeRateError
	if errors.As(err, &sameCurrencyExchangeRateErr) {
		problem.Type = problemBaseURL + "same-currency-exchange-rate"
		problem.Title = "Same Currency Exchange Rate"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr(sameCurrencyExchangeRateErr.Error())
		return problem, http.StatusBadRequest
	}

	// Unsupported currency
	var unsupportedCurrencyErr *domain.UnsupportedCurrencyError
	if errors.As(err, &unsupportedCurrencyErr) {
		problem.Type = problemBaseURL + "unsupported-currency"
		problem.Title = "Unsupported Currency"
		problem.Status = http.StatusBadRequest
		problem.Detail = ptr(unsupportedCurrencyErr.Error())
		return problem, http.StatusBadRequest
	}

	// Default: internal server error
	problem.Type = problemBaseURL + "internal-error"
	problem.Title = "Internal Server Error"
	problem.Status = http.StatusInternalServerError
	problem.Detail = ptr("An unexpected error occurred")
	return problem, http.StatusInternalServerError
}

// UnauthorizedError creates a ProblemDetails for unauthorized access.
func UnauthorizedError(instance string) ProblemDetails {
	return ProblemDetails{
		Type:     problemBaseURL + "unauthorized",
		Title:    "Unauthorized",
		Status:   http.StatusUnauthorized,
		Detail:   ptr("Authentication required"),
		Instance: ptr(instance),
	}
}

// ForbiddenError creates a ProblemDetails for forbidden access.
func ForbiddenError(instance string, detail string) ProblemDetails {
	return ProblemDetails{
		Type:     problemBaseURL + "forbidden",
		Title:    "Forbidden",
		Status:   http.StatusForbidden,
		Detail:   ptr(detail),
		Instance: ptr(instance),
	}
}

// ptr is a helper to create pointers to values.
func ptr[T any](v T) *T {
	return &v
}
