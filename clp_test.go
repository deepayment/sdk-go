package deepayment

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestCLPPaymentOptionalFieldsOnWire(t *testing.T) {
	for _, method := range []string{"WEBPAY", "KHIPU", "MACH"} {
		for _, extended := range []bool{false, true} {
			t.Run(method+map[bool]string{false: "/legacy", true: "/extended"}[extended], func(t *testing.T) {
				extra := &PaymentCustomerDocumentEmailExtra{
					DocumentType: "RUT", DocumentNumber: "12345678-9",
					CustomerName: "María Silva", CustomerEmail: "payer@example.com",
				}
				want := map[string]any{
					"documentType": "RUT", "documentNumber": "12345678-9",
					"customerName": "María Silva", "customerEmail": "payer@example.com",
				}
				if extended {
					extra.FirstName, extra.LastName, extra.CustomerPhone = "María", "Silva", "912345678"
					want["firstName"], want["lastName"], want["customerPhone"] = "María", "Silva", "912345678"
				}
				paymentMethod := PaymentMethod{Code: method}
				var branch string
				switch method {
				case "WEBPAY":
					paymentMethod.Webpay, branch = extra, "webpay"
				case "KHIPU":
					paymentMethod.Khipu, branch = extra, "khipu"
				case "MACH":
					paymentMethod.Mach, branch = extra, "mach"
				}
				k, captured := newTestKeys(t), &capture{}
				server := verifyingServer(t, k, `{}`, captured)
				defer server.Close()
				client := k.client(t, server.URL, server.Client())
				_, err := client.CreatePayment(context.Background(), &CreatePaymentReq{
					MerchantOrderNo: "clp-test-001", Currency: "CLP", Amount: "1000.00",
					WebhookUrl: "https://merchant.example.com/webhook", PaymentMethod: paymentMethod,
				})
				if err != nil {
					t.Fatal(err)
				}
				var request struct {
					PaymentMethod map[string]json.RawMessage `json:"paymentMethod"`
				}
				if err := json.Unmarshal(captured.openedBody, &request); err != nil {
					t.Fatal(err)
				}
				var got map[string]any
				if err := json.Unmarshal(request.PaymentMethod[branch], &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(want, got) {
					t.Fatalf("payment details = %v, want %v", got, want)
				}
			})
		}
	}
}
