package domain

type Transaction interface {
	Begin()
	Commit()
	Rollback()
}
