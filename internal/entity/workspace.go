package entity

type Workspace struct {
	ID          int64   `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	Type        string  `db:"type" json:"type"`
	HourlyRate  float64 `db:"hourly_rate" json:"hourly_rate"`
	Description string  `db:"description" json:"description"`
	Capacity    int     `db:"capacity" json:"capacity"`
}
