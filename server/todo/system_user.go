package todo

import (
	"github.com/shopspring/decimal"
	"server/model"
	"server/pkg/utils"
	"server/repository"
)

func systemUser() {
	db, _, _ := repository.Get("")
	count, err := db.SystemUser.Count()
	if err != nil {
		return
	}
	if count == 0 {
		salt := utils.RandStr(8, utils.AllDict)
		pwdHash, _ := utils.HashPassword("admin")
		_ = db.SystemUser.Create(&model.SystemUser{
			Base: model.Base{
				Id:   1,
				Code: "1",
			},
			Account:  "admin",
			Password: pwdHash,
			Salt:     salt,
			OtpKey:   "",
			Admin:    1,
			Amount:   decimal.NewFromInt(1000),
		})
	}
}
