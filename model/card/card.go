package card

// New creates a new card.
func New(name, cType, expireDate, number, cvc string) *Card {
	return &Card{
		Name:       name,
		Type:       cType,
		Number:     number,
		CVC:        cvc,
		ExpireDate: expireDate,
	}
}
