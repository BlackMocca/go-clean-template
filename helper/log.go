package helper

import (
	"git.innovasive.co.th/backend/helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

func Println(str string) {
	logger := cast.ToBool(helper.GetENV("APP_LOGGER", "false"))
	if logger {
		logrus.Debugln(str)
	}
}
