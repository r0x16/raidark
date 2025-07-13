package driver

import (
	"sync"

	"github.com/r0x16/Raidark/shared/datastore/domain"
	"gorm.io/gorm"
)

type GormTransaction struct {
	db               *gorm.DB
	tx               *gorm.DB
	IsInTransaction  bool
	TransactionMutex sync.Mutex
}

var _ domain.Transaction = &GormTransaction{}

func NewGormTransaction(db *gorm.DB) *GormTransaction {
	return &GormTransaction{db: db, IsInTransaction: false}
}

func (t *GormTransaction) Begin() {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if t.IsInTransaction {
		return
	}
	if t.tx != nil {
		return
	}

	t.tx = t.db.Begin()
	t.IsInTransaction = true
}

func (t *GormTransaction) Commit() {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if !t.IsInTransaction {
		return
	}

	t.tx.Commit()
	t.tx = nil
	t.IsInTransaction = false
}

func (t *GormTransaction) Rollback() {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if !t.IsInTransaction {
		return
	}

	t.tx.Rollback()
	t.tx = nil
	t.IsInTransaction = false
}
