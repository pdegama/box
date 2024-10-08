package server

var mailFwd ForwardBackend = nil // it is hold MailForward struct by config

type ForwardBackend interface {
	Init()
	ForwardMail(Email)
	ExistMailBox(string) bool
}

// this is globle and overridable
func SetMailFwdMethod(backend ForwardBackend) {
	mailFwd = backend
	mailFwd.Init()
}
