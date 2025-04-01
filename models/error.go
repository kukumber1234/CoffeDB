package models

type Error struct {
	Message string `json:"Error"`
	Status  int64  `json:"Status"`
}
