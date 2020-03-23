package app

type Payload struct {
	Id          int64    `json:"id"`
	Exp         int64    `json:"exp"`
	NameSurname string   `json:"name_surname"`
	Roles       []string `json:"roles"`
}
