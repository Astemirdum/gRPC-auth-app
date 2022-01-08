package entity

type User struct {
	Id       int    `json:"id" db:"id"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password_hash"`
}
