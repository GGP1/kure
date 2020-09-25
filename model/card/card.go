package card

// New returns a new card struct.
func New(name, cType, expireDate string, number, cvc int32) *Card {
	return &Card{
		Name:       name,
		Type:       cType,
		Number:     number,
		CVC:        cvc,
		ExpireDate: expireDate,
	}
}
