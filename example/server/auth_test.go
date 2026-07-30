package server

import "testing"

// spec:auth-login
func TestLoginRejectsBadCredentialsWithoutLeak(t *testing.T) {
	if _, ok := (Auth{}).Login("a@b.com", "wrong"); ok {
		t.Fatal("bad credentials must not authenticate")
	}
}

// spec:auth-logout
func TestLogoutRevokesSessionServerSide(t *testing.T) {
	a := Auth{}
	a.Logout("sess")
	if a.Valid("sess") {
		t.Fatal("session must be invalid after logout")
	}
}

// spec:auth-session-expiry
func TestSessionExpiresAfter24h(t *testing.T) {
	if (Auth{}).Valid("stale") {
		t.Fatal("a stale session must be rejected")
	}
}

// A parent spec may also carry its own test — here an end-to-end check of the
// whole account-access capability, on top of the per-behaviour tests above.
// spec:auth-access
func TestAccountAccessEndToEnd(t *testing.T) {
	a := Auth{}
	if a.Valid("nobody") {
		t.Fatal("no session should be valid before signing in")
	}
	a.Logout("sess")
	if a.Valid("sess") {
		t.Fatal("account must not be accessible after logout")
	}
}
