package domain

// Bank represents a bank entity (read-only lookup)
type Bank struct {
	id   int
	name string
	code string
}

func NewBank(id int, name, code string) Bank {
	return Bank{id: id, name: name, code: code}
}

func (b Bank) ID() int      { return b.id }
func (b Bank) Name() string { return b.name }
func (b Bank) Code() string { return b.code }
