// Package pgutil holds conversion helpers between Go values and pgtype values.
package pgutil

import (
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

// NumericFromFloatPtr builds a pgtype.Numeric from an optional float64.
// A nil pointer yields a NULL numeric.
func NumericFromFloatPtr(f *float64) pgtype.Numeric {
	var n pgtype.Numeric
	if f == nil {
		n.Valid = false
		return n
	}
	_ = n.Scan(formatFloat(*f))
	return n
}

// NumericToFloatPtr converts a pgtype.Numeric to an optional float64.
func NumericToFloatPtr(n pgtype.Numeric) *float64 {
	if !n.Valid || n.NaN {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}

// NumericFromInt builds an exact pgtype.Numeric from an int64.
func NumericFromInt(i int64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(i), Exp: 0, Valid: true}
}

func formatFloat(f float64) string {
	// strconv via big.Float keeps precision reasonable for currency/area values.
	return new(big.Float).SetFloat64(f).Text('f', 3)
}
