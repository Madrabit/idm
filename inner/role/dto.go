package role

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
