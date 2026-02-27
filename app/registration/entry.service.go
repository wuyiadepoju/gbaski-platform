package registration

type EntryService struct {
	repo *Repository
}

func NewEntryService() *EntryService {
	return &EntryService{NewRepository()}
}
