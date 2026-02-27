package setting

import "github.com/gbaski/gbaski-shared/repo"

type Repository struct {
	*repo.BaseRepository
}

func NewRepository() *Repository {
	return &Repository{repo.NewBaseRepository()}
}
