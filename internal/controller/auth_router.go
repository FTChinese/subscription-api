package controller

type AuthRouter struct {
	UserShared
}

func NewAuthRouter(shared UserShared) AuthRouter {
	return AuthRouter{
		UserShared: shared,
	}
}
