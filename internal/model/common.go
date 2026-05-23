package model

import "github.com/jmoiron/sqlx"

type PageResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"page_size"`
}

func NewPageResult(list interface{}, total int64, page, size int) *PageResult {
	return &PageResult{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	}
}

type DB struct {
	*sqlx.DB
}

func NewDB(db *sqlx.DB) *DB {
	return &DB{db}
}
