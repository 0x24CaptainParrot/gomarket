package services

type Service struct {
	AuthService Authorization

}

func NewService(deps Dependencies) *Service {
	return &Service{
		AuthService: deps.AuthService,
	}
}

type Dependencies struct {
	AuthService Authorization
}
