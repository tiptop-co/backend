package venue

type Venue struct {
	ID          string  `json:"id"           db:"id"`
	Name        string  `json:"name"         db:"name"`
	Address     string  `json:"address"      db:"address"`
	Description *string `json:"description"  db:"description"`
	BankAccount *string `json:"bank_account" db:"bank_account"`
	ManagerID   *string `json:"manager_id"   db:"manager_id"`
}

type UpdateInput struct {
	Name        string
	Address     string
	Description *string
	BankAccount *string
}
