package deepayment

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAggregateValidationVectors(t *testing.T) {
	raw, err := os.ReadFile("protocol/testdata/methods/002-aggregate.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Direction, Currency string
		Method                    json.RawMessage
		Valid                     bool
	}
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			if c.Direction == "payment" {
				var m PaymentMethod
				if err = json.Unmarshal(c.Method, &m); err == nil {
					err = validatePaymentMethod(c.Currency, m)
				}
			} else {
				var m PayoutMethod
				if err = json.Unmarshal(c.Method, &m); err == nil {
					err = validatePayoutMethod(c.Currency, m)
				}
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v error=%v", c.Valid, err)
			}
		})
	}
}
