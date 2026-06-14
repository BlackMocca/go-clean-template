package models

type Paginator[T any] struct {
	Page          int64 `json:"page"`
	PageSize      int64 `json:"page_size"`
	TotalPage     int64 `json:"total_page"`
	TotalDocument int64 `json:"total_documents"`
	Documents     T     `json:"documents"`
}

func (p *Paginator[T]) Limit() int64 {
	return p.PageSize
}

func (p *Paginator[T]) Offset() int64 {
	if p.Page <= 0 {
		p.Page = 1
	}

	if p.PageSize <= 0 {
		return 0
	}

	return (p.Page - 1) * p.PageSize
}

func (p *Paginator[T]) SetTotalDocument(totalDocument int64) {
	p.TotalDocument = totalDocument
	if p.PageSize <= 0 {
		p.TotalPage = 0
		return
	}

	p.TotalPage = (totalDocument + p.PageSize - 1) / p.PageSize
}

func (p *Paginator[T]) SetDocument(documents T) {
	p.Documents = documents
}
