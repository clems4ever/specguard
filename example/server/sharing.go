package server

type Sharing struct{}

func (Sharing) Invite(taskID, email string) error { return nil }

// Promote is unimplemented — see the draft spec sharing-permissions.
func (Sharing) Promote(taskID, email string) error { return nil }
