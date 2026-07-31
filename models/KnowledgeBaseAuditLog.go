package models

import (
	"time"
)

// KnowledgeBaseAuditLog menyimpan rekam jejak audit (audit trail) untuk setiap
// perubahan (Create, Update, Delete) pada basis pengetahuan (penyakit, pertanyaan, aturan).
type KnowledgeBaseAuditLog struct {
	LogId              int64     `gorm:"primaryKey;autoIncrement"          json:"log_id"`
	LogUID             string    `gorm:"type:varchar(255);uniqueIndex"     json:"log_uid"`
	OperatorUID        string    `gorm:"type:varchar(255);index"           json:"operator_uid"`  // pakar_uid atau admin_uid
	OperatorName       string    `gorm:"type:varchar(90)"                  json:"operator_name"`
	OperatorRole       string    `gorm:"type:varchar(30)"                  json:"operator_role"` // "pakar" | "admin"
	TabelTerdampak     string    `gorm:"type:varchar(50);index"            json:"tabel_terdampak"` // "penyakits" | "pertanyaans" | "aturans"
	RecordIdTerdampak  string    `gorm:"type:varchar(255);index"           json:"record_id_terdampak"` // UID record yang berubah
	Aksi               string    `gorm:"type:varchar(20)"                  json:"aksi"` // "CREATE" | "UPDATE" | "DELETE"
	DataSebelum        *string   `gorm:"type:jsonb"                        json:"data_sebelum"`
	DataSesudah        *string   `gorm:"type:jsonb"                        json:"data_sesudah"`
	CreatedAt          time.Time `gorm:"type:timestamp"                    json:"created_at"`
}
