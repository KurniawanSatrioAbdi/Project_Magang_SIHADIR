package models

type Presensi struct {
	ID        int `gorm:"primaryKey"`
	UserID    int
	Tanggal   string
	JamMasuk  string
	JamPulang string
	Status    string
	Latitude  string
	Longitude string
	Foto      string
}

func (Presensi) TableName() string {
	return "presensi"
}
