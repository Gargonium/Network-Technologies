package dns

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
	"time"
)

func ResolveDNSName(domain string) (string, error) {
	client := &dns.Client{}
	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, _, err := client.ExchangeContext(ctx, msg, "8.8.8.8:53") // Используем публичный DNS Google
	if err != nil {
		return "", err
	}

	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			return a.A.String(), nil
		}
	}
	return "", fmt.Errorf("не удалось найти IPv4 адрес для домена %s", domain)
}
