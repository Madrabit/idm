package employee

import "time"

type Entity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (e Entity) toResponse() Response {
	return Response(e)
}

type Response struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type NameRequest struct {
	Name string `json:"name" validate:"required,min=2,max=155"`
}

func (req *NameRequest) toEntity() Entity {
	return Entity{Name: req.Name}
}

type IdRequest struct {
	Id int64 `param:"id" validate:"gt=0"`
}

type IdsRequest struct {
	Ids []int64 `json:"ids" validate:"required,min=1,dive,gt=0"`
}

type PageRequest struct {
	PageSize   int64 `validate:"min=1,max=100"`
	PageNumber int64 `validate:"min=0"`
}

type PageKeySetRequest struct {
	LastId   int64 `validate:"min=0"`
	PageSize int64 `validate:"min=1,max=100"`
	IsNext   bool
}

type PageResponse struct {
	Result     []Entity `json:"result"`
	PageSize   int64    `json:"page_size" `
	PageNumber int64    `json:"page_number"`
	Total      int64    `json:"total"`
}

type PageKeySetResponse struct {
	Result []Entity `json:"result"`
	LastId int64    `json:"last_id"`
	Total  int64    `json:"total"`
}
