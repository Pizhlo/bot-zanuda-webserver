package auth

type AuthService struct {
	secretKey []byte
}

func New(cfg *config) *AuthService {
	return &AuthService{secretKey: cfg.secretKey}
}
