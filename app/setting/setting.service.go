package setting

type Service struct {
	repo *Repository
}

func NewService() *Service {
	return &Service{repo: NewRepository()}
}

type SettingHandler struct {
	*Service
}

func NewSettingHandler() *SettingHandler {
	return &SettingHandler{NewService()}
}
