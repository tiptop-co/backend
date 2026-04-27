package table_handler

import "github.com/tiptop-co/backend/internal/model/table"

type createTableRequest struct {
	Number int `json:"number"`
}

type tableByQRRequest struct {
	QRToken string `json:"qr_token"`
}

type tablesResponse struct {
	Tables []*table.Table `json:"tables"`
}
