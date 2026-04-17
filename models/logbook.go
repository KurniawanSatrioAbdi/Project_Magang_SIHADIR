package models

type Logbook struct {
	ID        uint
	UserID    uint
	Tanggal   string
	Kegiatan  string
	Deskripsi string
}

func (Logbook) TableName() string {
	return "logbook"
}
