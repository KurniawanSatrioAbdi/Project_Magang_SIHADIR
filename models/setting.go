package models

type Setting struct {
	ID             uint
	JamMasuk       string
	JamPulang      string
	Toleransi      string
	JamMasukJumat  string
	JamPulangJumat string
	ToleransiJumat string
	KantorLat      string
	KantorLng      string
	RadiusMeter    int
}

func (Setting) TableName() string {
	return "settings"
}
