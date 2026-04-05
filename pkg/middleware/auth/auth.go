package auth

type Middleware struct{}

func New() *Middleware {
	return &Middleware{}
}
