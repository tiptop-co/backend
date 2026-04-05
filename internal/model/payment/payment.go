package payment

// НЕ ИСПОЛЬЗОВАТЬ
// Все содержимое данного файла является примером!

const (
	StatusPending   = "pending"
	StatusCancelled = "cancelled"
	StatusRefunded  = "refunded"
	StatusPaid      = "paid"
)

type Payment struct {
	TransactionID int64  `db:"transaction_id"`
	Status        string `db:"status"`
}
