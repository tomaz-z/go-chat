package models

import "github.com/oklog/ulid/v2"

const BearerToken = "Bearer"

type User struct {
	Name string
}

type Token struct {
	Value ulid.ULID
}

type Message struct {
	Author string
	Value  string
}
