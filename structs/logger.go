package structs

type JwtSessionPayload struct {
	TID  string `json:"tid"  binding:"required"`
	Type string `json:"type" binding:"required"`
	RID  string `json:"rid"  binding:"required"`
	SID  string `json:"sid"  binding:"required"`
}
