package helper

import (
	"fmt"

	"github.com/BlackMocca/go-clean-template/config"
)

func Println(str string) {
	if config.APP_LOGGER {
		fmt.Println(str)
	}
}
