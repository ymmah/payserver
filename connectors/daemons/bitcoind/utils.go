package bitcoind

import (
	"math/big"

	"strconv"

	"github.com/btcsuite/btcutil"
	"github.com/shopspring/decimal"
	"strings"
	"github.com/bitlum/connector/connectors/assets/bitcoin"
	"github.com/bitlum/connector/connectors/assets/litecoin"
	"github.com/bitlum/connector/connectors/assets/bitcoincash"
	"github.com/bitlum/connector/connectors/assets/dash"
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcd/chaincfg"
)

var satoshiPerBitcoin = decimal.New(btcutil.SatoshiPerBitcoin, 0)

func decAmount2Sat(amount decimal.Decimal) btcutil.Amount {
	// If we would try to convert amount in float representation than it
	// could lead to precious error, for that reason convert in manually rather
	// than using btcutil.NewAmount().
	amtStr := amount.Mul(satoshiPerBitcoin).String()
	a, _ := strconv.ParseInt(amtStr, 10, 64)
	return btcutil.Amount(a)
}

func sat2DecAmount(amount btcutil.Amount) decimal.Decimal {
	amt := decimal.NewFromBigInt(big.NewInt(int64(amount)), 0)
	return amt.Div(satoshiPerBitcoin)
}

func printAmount(a btcutil.Amount) string {
	u := btcutil.AmountBTC
	return strconv.FormatFloat(a.ToUnit(u), 'f', -int(u+8), 64)
}

func isProperNet(desiredNet, actualNet string) bool {
	// Handle the case of different simulation networks names
	if desiredNet == "simnet" && actualNet == "regtest" {
		return true
	}

	// Handle the case of different testnet networks names
	if desiredNet == "testnet" && actualNet == "test" {
		return true
	}

	// Handle the case of different mainnet networks names
	if desiredNet == "mainnet" && actualNet == "main" {
		return true
	}

	return desiredNet == actualNet
}