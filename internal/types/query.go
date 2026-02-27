package types

import "time"

type SortQuery struct {
	Page       uint   `json:"page" form:"page"`
	SortBy     string `json:"sort_by" form:"sort_by"`
	SortOrder  string `json:"sort_order" form:"sort_order"`
	Size       int    `json:"size" form:"size"`
	DeletedAt  string `json:"deleted_at" form:"deleted_at" example:"show|only|"`
	LikeSearch bool   `json:"like_search" form:"like_search"`
}

type ThemeFilter struct {
	SortQuery
	Name     string
	Names    []string
	UserId   uint
	UserTGId int64
}

type TaskFilter struct {
	SortQuery
	Name       string
	Names      []string
	ThemeName  string
	ThemeNames []string
	UserId     uint
	UserTGId   int64
	Status     int
	Priority   int
	Deadline   time.Time // todo тип значения для поиска по времени ?
}
