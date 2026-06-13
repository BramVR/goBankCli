package authflow

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestListenLoopbackPreservesRequestedHost(t *testing.T) {
	listener, callbackAddress, err := listenLoopback("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if !strings.HasPrefix(callbackAddress, "localhost:") {
		t.Fatalf("callback address = %q, want localhost host", callbackAddress)
	}
}

func TestListenLoopbackRejectsNonLoopbackHost(t *testing.T) {
	if _, _, err := listenLoopback("192.0.2.10:8787"); err == nil || !strings.Contains(err.Error(), "must be loopback") {
		t.Fatalf("listenLoopback error = %v, want loopback rejection", err)
	}
}

func TestLocalCallbackIgnoresErrorWithWrongState(t *testing.T) {
	listener, callbackAddress, err := listenLoopback("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan callbackResult, 1)
	go func() {
		callback, err := waitForLocalCallback(context.Background(), listener, "/enablebanking/callback", "expected-state", 2*time.Second, localCallbackTLSConfig{})
		done <- callbackResult{callback: callback, err: err}
	}()

	resp, err := http.Get(localCallbackURL(callbackAddress, "/enablebanking/callback", false) + "?error=access_denied&state=wrong-state")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("wrong-state error status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	select {
	case item := <-done:
		t.Fatalf("callback completed early: %#v", item)
	case <-time.After(50 * time.Millisecond):
	}

	resp, err = http.Get(localCallbackURL(callbackAddress, "/enablebanking/callback", false) + "?code=callback-code&state=expected-state")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("valid callback status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	select {
	case item := <-done:
		if item.err != nil {
			t.Fatal(item.err)
		}
		if item.callback.Code != "callback-code" || item.callback.State != "expected-state" {
			t.Fatalf("callback = %#v", item.callback)
		}
	case <-time.After(time.Second):
		t.Fatal("callback did not complete")
	}
}

func TestLocalHTTPSCallbackAcceptsRequest(t *testing.T) {
	listener, callbackAddress, err := listenLoopback("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan callbackResult, 1)
	go func() {
		callback, err := waitForLocalCallback(context.Background(), listener, "/enablebanking/callback", "expected-state", 2*time.Second, localCallbackTLSConfig{Enabled: true})
		done <- callbackResult{callback: callback, err: err}
	}()

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get(localCallbackURL(callbackAddress, "/enablebanking/callback", true) + "?code=callback-code&state=expected-state")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("valid callback status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	select {
	case item := <-done:
		if item.err != nil {
			t.Fatal(item.err)
		}
		if item.callback.Code != "callback-code" || item.callback.State != "expected-state" {
			t.Fatalf("callback = %#v", item.callback)
		}
	case <-time.After(time.Second):
		t.Fatal("callback did not complete")
	}
}
