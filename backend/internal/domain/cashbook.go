package domain

import "github.com/google/uuid"

var (
	CashbookUserID = UserID(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	CashbookUSD    = AccountID(uuid.MustParse("00000000-0000-0000-0000-000000000010"))
	CashbookEUR    = AccountID(uuid.MustParse("00000000-0000-0000-0000-000000000011"))
)

func GetCashbookAccount(currency Currency) AccountID {
	switch currency {
	case CurrencyUSD:
		return CashbookUSD
	case CurrencyEUR:
		return CashbookEUR
	default:
		return CashbookUSD
	}
}
