package models

type User struct {
	Id       int    `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password_hash"`
}

type UserLog struct {
	Id        int32  `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Timestamp int64  `json:"timestamp"`
}
