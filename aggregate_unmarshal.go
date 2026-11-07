package deepayment

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UnmarshalJSON rejects aggregate payloads before the typed decoder can silently
// discard unknown fields. Ordinary method decoding retains its existing shape.
func (m *PaymentMethod) UnmarshalJSON(raw []byte) error {
	type plain PaymentMethod
	var value plain
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	if strings.TrimSpace(value.Code) == AggregatePaymentCode {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return err
		}
		if len(fields) != 1 {
			return fmt.Errorf("%w: aggregate payment is code-only", ErrConflictingMethodExtra)
		}
	}
	*m = PaymentMethod(value)
	return nil
}
