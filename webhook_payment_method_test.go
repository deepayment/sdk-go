package deepayment

import (
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

	merchantauth "github.com/deepayment/sdk-go/internal/merchantauth"
)

func TestPaymentWebhookMethodVectors(t *testing.T) {
	for _, tc := range []struct{ file, method, status string }{
		{"001-payment-succeeded.json", "", StatusSucceeded},
		{"004-payment-method.json", "PSE", StatusSucceeded},
		{"005-payment-method-unselected.json", "ALL", StatusExpired},
	} {
		t.Run(tc.file, func(t *testing.T) {
			v := readProtocolVector[webhookVector](t, "webhook/"+tc.file)
			k := newTestKeys(t)
			c := k.client(t, "https://api.example.com")
			c.now = func() time.Time { return time.Unix(1787803300, 0) }
			publicKey, err := base64.StdEncoding.DecodeString(v.Key.PlatformWebhookPublicKeyB64)
			if err != nil {
				t.Fatal(err)
			}
			c.platformWebhookPublicKeys = map[string]ed25519.PublicKey{v.Key.PlatformWebhookKeyID: publicKey}
			got, err := c.ParsePaymentWebhook(newWebhookVectorRequest(t, v))
			if err != nil {
				t.Fatal(err)
			}
			if got.PaymentMethod != tc.method || got.Status != tc.status {
				t.Fatalf("method=%q status=%q, want %q %q", got.PaymentMethod, got.Status, tc.method, tc.status)
			}
			if tc.method == "" {
				return
			}
			for _, refreshDigest := range []bool{false, true} {
				body := []byte(strings.Replace(v.Body, `"paymentMethod":"`+tc.method+`"`, `"paymentMethod":"NEQUI"`, 1))
				r := newWebhookVectorRequest(t, v)
				r.Body = io.NopCloser(strings.NewReader(string(body)))
				if refreshDigest {
					r.Header.Set(merchantauth.HeaderContentDigest, merchantauth.ContentDigestSHA256(body))
				}
				if _, err := c.ParsePaymentWebhook(r); err == nil {
					t.Fatalf("altered method accepted, refreshed digest=%v", refreshDigest)
				}
			}
		})
	}
}
