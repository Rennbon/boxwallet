package token

import (
	"github.com/Rennbon/boxwallet/bccoin"
	"github.com/Rennbon/boxwallet/bccore"
)

type TokenLawer interface {
	GetTokenInfo(contract bccore.Token) (*bccoin.CoinInfo, error)
}
