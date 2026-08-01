package adapters

import "testing"

func TestNewHTTPClientHasTimeout(t *testing.T) {
	client := NewHTTPClient()
	if client.Timeout != defaultHTTPTimeout {
		t.Fatalf("got timeout %s, want %s", client.Timeout, defaultHTTPTimeout)
	}
}

func TestRemoteAdaptersUseInjectedHTTPClient(t *testing.T) {
	client := NewHTTPClient()
	getter := NewRemoteGetterWithClient(true, client)
	if getter.client != client {
		t.Fatal("remote getter does not use the injected HTTP client")
	}

	poster := NewRemotePosterWithClient(client)
	if poster.client != client {
		t.Fatal("remote poster does not use the injected HTTP client")
	}
}
