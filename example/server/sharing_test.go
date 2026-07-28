package server

import "testing"

// spec:sharing-invite
func TestInviteQueuesUnknownEmail(t *testing.T) {
	if err := (Sharing{}).Invite("t1", "new@user.com"); err != nil {
		t.Fatal(err)
	}
}
