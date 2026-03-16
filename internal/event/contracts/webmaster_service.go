package contracts

// WebmasterService defines the contract for webmaster operations
type WebmasterService interface {
	AddEventURL(url string) error
}
