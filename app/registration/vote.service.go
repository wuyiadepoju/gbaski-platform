package registration

type VoteService struct {
	repo *Repository
}

func NewVoteService() *VoteService {
	return &VoteService{NewRepository()}
}
