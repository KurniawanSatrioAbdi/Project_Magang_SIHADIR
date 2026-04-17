package models

type Izin struct {
	ID         uint
	UserID     uint
	Tanggal    string
	Jenis      string
	Keterangan string
	FileSurat  string // path file surat, kosong jika tidak upload
	Status     string
}

func (Izin) TableName() string {
	return "izin"
}
