package server

// Auth is a stand-in for the example project's authentication logic. The real
// behaviour is described by the specs in specs/auth and pinned by auth_test.go.
type Auth struct{}

func (Auth) Login(email, password string) (session string, ok bool) { return "", false }
func (Auth) Logout(session string)                                  {}
func (Auth) Valid(session string) bool                              { return false }
