package models

type Paging struct {
	Page      int      `json:"page" form:"page"`
	Limit     int      `json:"limit" form:"limit"`
	Total     int64    `json:"total" form:"-"`
	Title     string   `json:"title" form:"title"`
	StartDate string   `json:"start_date" form:"start_date"`
	EndDate   string   `json:"end_date" form:"end_date"`
	Status    []string `json:"status" form:"status"`
}

func (p *Paging) Process() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 || p.Limit > 100 {
		p.Limit = 10
	}
	if p.Page <= 0 || p.Limit > 100 {
		p.Limit = 10
	}
}
func (p *Paging) OffSet() int {
	return (p.Page - 1) * p.Limit
}
