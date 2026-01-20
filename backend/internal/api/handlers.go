package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/shopspring/decimal"

	"minibankingplatform/internal/domain"
	"minibankingplatform/internal/service"
)

// APIHandler implements the StrictServerInterface.
type APIHandler struct {
	service *service.Service
}

// NewAPIHandler creates a new APIHandler with the given service.
func NewAPIHandler(svc *service.Service) *APIHandler {
	return &APIHandler{service: svc}
}

// Compile-time check that APIHandler implements StrictServerInterface.
var _ StrictServerInterface = (*APIHandler)(nil)

// Register handles user registration.
func (h *APIHandler) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	// Validate request
	if err := ValidateStruct(request.Body); err != nil {
		problem, _ := MapError(err, "/auth/register")
		return Register400ApplicationProblemPlusJSONResponse(problem), nil
	}

	cmd := &service.RegisterCommand{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}

	result, err := h.service.Register(ctx, cmd)
	if err != nil {
		return h.mapRegisterError(err)
	}

	return Register201JSONResponse{
		UserId: ptr(openapi_types.UUID(result.UserID)),
		Email:  ptr(openapi_types.Email(result.Email)),
		Token:  ptr(result.Token),
	}, nil
}

func (h *APIHandler) mapRegisterError(err error) (RegisterResponseObject, error) {
	var userExistsErr *domain.UserAlreadyExistsError
	if errors.As(err, &userExistsErr) {
		problem, _ := MapError(err, "/auth/register")
		return Register409ApplicationProblemPlusJSONResponse(problem), nil
	}

	problem, _ := MapError(err, "/auth/register")
	return Register400ApplicationProblemPlusJSONResponse(problem), nil
}

// Login handles user authentication.
func (h *APIHandler) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
	// Validate request
	if err := ValidateStruct(request.Body); err != nil {
		problem, _ := MapError(err, "/auth/login")
		return Login400ApplicationProblemPlusJSONResponse(problem), nil
	}

	cmd := &service.LoginCommand{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}

	result, err := h.service.Login(ctx, cmd)
	if err != nil {
		var invalidCredsErr *domain.InvalidCredentialsError
		if errors.As(err, &invalidCredsErr) {
			problem, _ := MapError(err, "/auth/login")
			return Login401ApplicationProblemPlusJSONResponse(problem), nil
		}
		problem, _ := MapError(err, "/auth/login")
		return Login400ApplicationProblemPlusJSONResponse(problem), nil
	}

	return Login200JSONResponse{
		UserId: ptr(openapi_types.UUID(result.UserID)),
		Email:  ptr(openapi_types.Email(result.Email)),
		Token:  ptr(result.Token),
	}, nil
}

// GetCurrentUser returns information about the authenticated user.
func (h *APIHandler) GetCurrentUser(ctx context.Context, _ GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error) {
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return GetCurrentUser401ApplicationProblemPlusJSONResponse(UnauthorizedError("/auth/me")), nil
	}

	return GetCurrentUser200JSONResponse{
		UserId: ptr(openapi_types.UUID(claims.UserID)),
		Email:  ptr(openapi_types.Email(claims.Email)),
	}, nil
}

// ListAccounts returns all accounts for the authenticated user.
func (h *APIHandler) ListAccounts(ctx context.Context, _ ListAccountsRequestObject) (ListAccountsResponseObject, error) {
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return ListAccounts401ApplicationProblemPlusJSONResponse(UnauthorizedError("/accounts")), nil
	}

	accounts, err := h.service.GetUserAccounts(ctx, domain.UserID(userID))
	if err != nil {
		problem, _ := MapError(err, "/accounts")
		return ListAccounts401ApplicationProblemPlusJSONResponse(problem), nil
	}

	response := make([]Account, len(accounts))
	for i, acc := range accounts {
		response[i] = domainAccountToAPI(acc)
	}

	return ListAccounts200JSONResponse(response), nil
}

// GetAccountBalance returns the balance of a specific account.
func (h *APIHandler) GetAccountBalance(ctx context.Context, request GetAccountBalanceRequestObject) (GetAccountBalanceResponseObject, error) {
	_, err := UserIDFromContext(ctx)
	if err != nil {
		return GetAccountBalance401ApplicationProblemPlusJSONResponse(UnauthorizedError("/accounts/" + request.AccountId.String() + "/balance")), nil
	}

	balance, err := h.service.GetAccountBalance(ctx, domain.AccountID(request.AccountId))
	if err != nil {
		var notFoundErr *domain.AccountNotFoundError
		if errors.As(err, &notFoundErr) {
			problem, _ := MapError(err, "/accounts/"+request.AccountId.String()+"/balance")
			return GetAccountBalance404ApplicationProblemPlusJSONResponse(problem), nil
		}
		problem, _ := MapError(err, "/accounts/"+request.AccountId.String()+"/balance")
		return GetAccountBalance401ApplicationProblemPlusJSONResponse(problem), nil
	}

	return GetAccountBalance200JSONResponse{
		AccountId: ptr(request.AccountId),
		Balance:   domainMoneyToAPI(balance),
	}, nil
}

// Transfer handles money transfer between accounts.
func (h *APIHandler) Transfer(ctx context.Context, request TransferRequestObject) (TransferResponseObject, error) {
	_, err := UserIDFromContext(ctx)
	if err != nil {
		return Transfer401ApplicationProblemPlusJSONResponse(UnauthorizedError("/transactions/transfer")), nil
	}

	// Validate request
	if err := ValidateStruct(request.Body); err != nil {
		problem, _ := MapError(err, "/transactions/transfer")
		return Transfer400ApplicationProblemPlusJSONResponse(problem), nil
	}

	now := time.Now()
	cmd, err := service.NewTransferCommand(
		uuid.UUID(request.Body.FromAccountId),
		uuid.UUID(request.Body.ToAccountId),
		request.Body.Amount,
		string(request.Body.Currency),
		now,
	)
	if err != nil {
		problem, _ := MapError(err, "/transactions/transfer")
		return Transfer400ApplicationProblemPlusJSONResponse(problem), nil
	}

	err = h.service.Transfer(ctx, cmd)
	if err != nil {
		return h.mapTransferError(err, request.Body.FromAccountId, request.Body.ToAccountId)
	}

	return Transfer200JSONResponse{
		TransactionId: ptr(openapi_types.UUID(uuid.New())), // Note: service doesn't return transaction ID yet
		FromAccountId: ptr(request.Body.FromAccountId),
		ToAccountId:   ptr(request.Body.ToAccountId),
		Amount: &Money{
			Amount:   ptr(request.Body.Amount),
			Currency: ptr(request.Body.Currency),
		},
		Timestamp: ptr(now),
	}, nil
}

func (h *APIHandler) mapTransferError(err error, fromAccount, toAccount openapi_types.UUID) (TransferResponseObject, error) {
	var notFoundErr *domain.AccountNotFoundError
	if errors.As(err, &notFoundErr) {
		problem, _ := MapError(err, "/transactions/transfer")
		return Transfer404ApplicationProblemPlusJSONResponse(problem), nil
	}

	problem, _ := MapError(err, "/transactions/transfer")
	return Transfer400ApplicationProblemPlusJSONResponse(problem), nil
}

// Exchange handles currency exchange between user's accounts.
func (h *APIHandler) Exchange(ctx context.Context, request ExchangeRequestObject) (ExchangeResponseObject, error) {
	_, err := UserIDFromContext(ctx)
	if err != nil {
		return Exchange401ApplicationProblemPlusJSONResponse(UnauthorizedError("/transactions/exchange")), nil
	}

	// Validate request
	if err := ValidateStruct(request.Body); err != nil {
		problem, _ := MapError(err, "/transactions/exchange")
		return Exchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	// We need to get the source account currency for the exchange command
	// First, get the source account to know its currency
	sourceBalance, err := h.service.GetAccountBalance(ctx, domain.AccountID(request.Body.SourceAccountId))
	if err != nil {
		var notFoundErr *domain.AccountNotFoundError
		if errors.As(err, &notFoundErr) {
			problem, _ := MapError(err, "/transactions/exchange")
			return Exchange404ApplicationProblemPlusJSONResponse(problem), nil
		}
		problem, _ := MapError(err, "/transactions/exchange")
		return Exchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	now := time.Now()
	cmd, err := service.NewExchangeCommand(
		uuid.UUID(request.Body.SourceAccountId),
		uuid.UUID(request.Body.TargetAccountId),
		request.Body.Amount,
		string(sourceBalance.Currency()),
		now,
	)
	if err != nil {
		problem, _ := MapError(err, "/transactions/exchange")
		return Exchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	err = h.service.Exchange(ctx, cmd)
	if err != nil {
		return h.mapExchangeError(err)
	}

	// Note: Service doesn't return exchange details, so we return a basic response
	return Exchange200JSONResponse{
		TransactionId:   ptr(openapi_types.UUID(uuid.New())),
		SourceAccountId: ptr(request.Body.SourceAccountId),
		TargetAccountId: ptr(request.Body.TargetAccountId),
		SourceAmount: &Money{
			Amount:   ptr(request.Body.Amount),
			Currency: ptr(Currency(sourceBalance.Currency())),
		},
		Timestamp: ptr(now),
	}, nil
}

func (h *APIHandler) mapExchangeError(err error) (ExchangeResponseObject, error) {
	var notFoundErr *domain.AccountNotFoundError
	if errors.As(err, &notFoundErr) {
		problem, _ := MapError(err, "/transactions/exchange")
		return Exchange404ApplicationProblemPlusJSONResponse(problem), nil
	}

	problem, _ := MapError(err, "/transactions/exchange")
	return Exchange400ApplicationProblemPlusJSONResponse(problem), nil
}

// CalculateExchange calculates the exchange amount without executing the exchange.
func (h *APIHandler) CalculateExchange(ctx context.Context, request CalculateExchangeRequestObject) (CalculateExchangeResponseObject, error) {
	_, err := UserIDFromContext(ctx)
	if err != nil {
		return CalculateExchange401ApplicationProblemPlusJSONResponse(UnauthorizedError("/transactions/exchange/calculate")), nil
	}

	// Map currencies
	sourceCurrency, err := mapAPICurrencyToDomain(request.Params.SourceCurrency)
	if err != nil {
		problem, _ := MapError(err, "/transactions/exchange/calculate")
		return CalculateExchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	targetCurrency, err := mapAPICurrencyToDomain(request.Params.TargetCurrency)
	if err != nil {
		problem, _ := MapError(err, "/transactions/exchange/calculate")
		return CalculateExchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	// Parse amount
	decimalAmount, err := parseDecimalAmount(request.Params.Amount)
	if err != nil {
		problem := ProblemDetails{
			Type:     problemBaseURL + "validation-error",
			Title:    "Validation Error",
			Status:   http.StatusBadRequest,
			Detail:   ptr("Invalid amount format"),
			Instance: ptr("/transactions/exchange/calculate"),
		}
		return CalculateExchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	// Create source money
	sourceAmount, err := domain.NewMoney(decimalAmount, sourceCurrency)
	if err != nil {
		problem, _ := MapError(err, "/transactions/exchange/calculate")
		return CalculateExchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	// Calculate exchange
	result, err := h.service.CalculateExchangeAmount(sourceAmount, targetCurrency)
	if err != nil {
		problem, _ := MapError(err, "/transactions/exchange/calculate")
		return CalculateExchange400ApplicationProblemPlusJSONResponse(problem), nil
	}

	return CalculateExchange200JSONResponse{
		SourceAmount: &Money{
			Amount:   ptr(result.SourceAmount.Amount.String()),
			Currency: ptr(Currency(result.SourceAmount.Currency)),
		},
		TargetAmount: &Money{
			Amount:   ptr(result.TargetAmount.Amount.String()),
			Currency: ptr(Currency(result.TargetAmount.Currency)),
		},
		ExchangeRate: &struct {
			Rate           *string   `json:"rate,omitempty"`
			SourceCurrency *Currency `json:"sourceCurrency,omitempty"`
			TargetCurrency *Currency `json:"targetCurrency,omitempty"`
		}{
			Rate:           ptr(result.ExchangeRate.Rate().String()),
			SourceCurrency: ptr(Currency(result.ExchangeRate.From())),
			TargetCurrency: ptr(Currency(result.ExchangeRate.To())),
		},
	}, nil
}

// ListTransactions returns a paginated list of transactions.
func (h *APIHandler) ListTransactions(ctx context.Context, request ListTransactionsRequestObject) (ListTransactionsResponseObject, error) {
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return ListTransactions401ApplicationProblemPlusJSONResponse(UnauthorizedError("/transactions")), nil
	}

	// Default pagination
	page := 1
	limit := 20
	if request.Params.Page != nil && *request.Params.Page > 0 {
		page = *request.Params.Page
	}
	if request.Params.Limit != nil && *request.Params.Limit > 0 {
		limit = *request.Params.Limit
		if limit > 100 {
			limit = 100
		}
	}

	offset := (page - 1) * limit

	// Map transaction type filter
	var txType *domain.TransactionType
	if request.Params.Type != nil {
		domainType, err := mapAPITransactionTypeToDomain(*request.Params.Type)
		if err == nil {
			txType = &domainType
		}
	}

	cmd := &service.GetTransactionsCommand{
		UserID:          domain.UserID(userID),
		TransactionType: txType,
		Limit:           limit,
		Offset:          offset,
	}

	result, err := h.service.GetTransactions(ctx, cmd)
	if err != nil {
		problem, _ := MapError(err, "/transactions")
		return ListTransactions401ApplicationProblemPlusJSONResponse(problem), nil
	}

	// Map transactions
	transactions := make([]Transaction, len(result.Transactions))
	for i, tx := range result.Transactions {
		transactions[i] = domainTransactionToAPI(tx)
	}

	// Calculate pagination
	totalPages := (result.Total + limit - 1) / limit

	return ListTransactions200JSONResponse{
		Transactions: &transactions,
		Pagination: &Pagination{
			Total:      ptr(result.Total),
			Page:       ptr(page),
			Limit:      ptr(limit),
			TotalPages: ptr(totalPages),
		},
	}, nil
}

// Reconcile performs a reconciliation check and returns the report.
func (h *APIHandler) Reconcile(ctx context.Context, _ ReconcileRequestObject) (ReconcileResponseObject, error) {
	_, err := UserIDFromContext(ctx)
	if err != nil {
		return Reconcile401ApplicationProblemPlusJSONResponse(UnauthorizedError("/system/reconcile")), nil
	}

	report, err := h.service.Reconcile(ctx)
	if err != nil {
		problem, _ := MapError(err, "/system/reconcile")
		return Reconcile401ApplicationProblemPlusJSONResponse(problem), nil
	}

	// Map ledger balances
	ledgerBalances := make([]LedgerCurrencyStatus, len(report.LedgerBalances))
	for i, lb := range report.LedgerBalances {
		ledgerBalances[i] = LedgerCurrencyStatus{
			Currency:   ptr(Currency(lb.Currency)),
			TotalSum:   ptr(lb.TotalSum.String()),
			IsBalanced: ptr(lb.IsBalanced),
		}
	}

	// Map account mismatches
	accountMismatches := make([]AccountMismatch, len(report.AccountMismatches))
	for i, am := range report.AccountMismatches {
		accountMismatches[i] = AccountMismatch{
			AccountId:      ptr(openapi_types.UUID(am.AccountID)),
			Currency:       ptr(Currency(am.Currency)),
			AccountBalance: ptr(am.AccountBalance.String()),
			LedgerBalance:  ptr(am.LedgerBalance.String()),
			Difference:     ptr(am.Difference.String()),
		}
	}

	return Reconcile200JSONResponse{
		Timestamp:            ptr(report.Timestamp),
		IsConsistent:         ptr(report.IsConsistent),
		LedgerBalances:       &ledgerBalances,
		AccountMismatches:    &accountMismatches,
		TotalAccountsChecked: ptr(report.TotalAccountsChecked),
	}, nil
}

// Helper functions

func domainAccountToAPI(acc *domain.Account) Account {
	return Account{
		Id:      ptr(openapi_types.UUID(acc.ID())),
		UserId:  ptr(openapi_types.UUID(acc.UserID())),
		Balance: domainMoneyToAPI(acc.Balance()),
	}
}

func domainMoneyToAPI(m domain.Money) *Money {
	return &Money{
		Amount:   ptr(m.Amount().String()),
		Currency: ptr(Currency(m.Currency())),
	}
}

func domainTransactionToAPI(tx *domain.TransactionWithDetails) Transaction {
	result := Transaction{
		Id:        ptr(openapi_types.UUID(tx.Transaction().ID())),
		Type:      ptr(TransactionType(tx.Transaction().Type())),
		AccountId: ptr(openapi_types.UUID(tx.Transaction().Account())),
		Timestamp: ptr(tx.Transaction().Time()),
	}

	// Map transfer details if present
	if td := tx.TransferDetails(); td != nil {
		result.TransferDetails = &TransferDetails{
			Id:                 ptr(openapi_types.UUID(td.ID())),
			RecipientAccountId: ptr(openapi_types.UUID(td.RecipientAccount())),
			Amount: &Money{
				Amount:   ptr(td.Amount().Amount().String()),
				Currency: ptr(Currency(td.Amount().Currency())),
			},
		}
	}

	// Map exchange details if present
	if ed := tx.ExchangeDetails(); ed != nil {
		result.ExchangeDetails = &ExchangeDetails{
			Id:              ptr(openapi_types.UUID(ed.ID())),
			SourceAccountId: ptr(openapi_types.UUID(ed.SourceAccount())),
			TargetAccountId: ptr(openapi_types.UUID(ed.TargetAccount())),
			SourceAmount: &Money{
				Amount:   ptr(ed.SourceAmount().Amount().String()),
				Currency: ptr(Currency(ed.SourceAmount().Currency())),
			},
			TargetAmount: &Money{
				Amount:   ptr(ed.TargetAmount().Amount().String()),
				Currency: ptr(Currency(ed.TargetAmount().Currency())),
			},
			ExchangeRate: ptr(ed.ExchangeRate().String()),
		}
	}

	return result
}

func mapAPICurrencyToDomain(c Currency) (domain.Currency, error) {
	switch c {
	case USD:
		return domain.CurrencyUSD, nil
	case EUR:
		return domain.CurrencyEUR, nil
	default:
		return "", domain.NewUnsupportedCurrencyError(domain.Currency(c))
	}
}

func mapAPITransactionTypeToDomain(t TransactionType) (domain.TransactionType, error) {
	switch t {
	case Transfer:
		return domain.TransactionTypeTransfer, nil
	case Exchange:
		return domain.TransactionTypeExchange, nil
	case Deposit:
		return domain.TransactionTypeDeposit, nil
	case Withdrawal:
		return domain.TransactionTypeWithdrawal, nil
	default:
		return "", errors.New("unknown transaction type")
	}
}

func parseDecimalAmount(amount string) (decimal.Decimal, error) {
	return decimal.NewFromString(amount)
}
