package main

import (
	"context"

	"github.com/pewssh/gosdk/core/tokenrate"
)

func getUSDRate(symbol string) (float64, error) {
	return tokenrate.GetUSD(context.TODO(), symbol)
}
