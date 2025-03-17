package middlewares

import "time"

type Token struct {
	Value  string    `json:"value"  binding:"required"`
	Expiry time.Time `json:"expiry" binding:"required"`
}

type JwtAuthPayload struct {
	TID  string `json:"tid"  binding:"required"`
	Type string `json:"type" binding:"required"`
}

type JwtSessionPayload struct {
	TID  string `json:"tid"  binding:"required"`
	Type string `json:"type" binding:"required"`
	RID  string `json:"rid"  binding:"required"`
	SID  string `json:"sid"  binding:"required"`
}
