package employee

import "time"

type Entity struct {
	Id       int64     `db:"id"`
	Name     string    `db:"name"`
	CreateAt time.Time `db:"created_at"`
	UpdateAt time.Time `db:"updated_at"`
}

func (e Entity) toResponse() Response {
	return Response(e)
}

type Response struct {
	Id       int64     `json:"id"`
	Name     string    `json:"name"`
	CreateAt time.Time `json:"created_at"`
	UpdateAt time.Time `json:"updated_at"`
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
	Ids []int64 `param:"ids" validate:"min=1,dive,gt=0"`
}
