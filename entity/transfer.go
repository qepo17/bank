package entity

type Transfer struct {
	Model
	FromAccountID uint64
	ToAccountID   uint64
}
