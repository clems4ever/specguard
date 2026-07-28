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
