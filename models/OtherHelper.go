package models

import "github.com/go-basic/uuid"

//获取UUId
func GetUUid() (uu string) {
	uu = uuid.New()
	return
}
