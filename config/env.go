package config

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"git.innovasive.co.th/backend/helper"
	"github.com/spf13/cast"
)

var (
	ROOT_PATH    string
	APP_LOGGER   = cast.ToBool(helper.GetENV("APP_LOGGER", "true"))
	APP_PORT     = helper.GetENV("APP_PORT", "3000")
	JWT_SECRET   = helper.GetENV("JWT_SECRET", "test")
	GRPC_PORT    = helper.GetENV("GRPC_PORT", "3100")
	GRPC_TIMEOUT = cast.ToInt(helper.GetENV("GRPC_TIMEOUT", "120"))

	SENTRY_DSN = helper.GetENV("SENTRY_DSN", "")

	PSQL_DATABASE_URL = helper.GetENV("PSQL_DATABASE_URL", "postgres://postgres:postgres@psql_db:5432/app_example?sslmode=disable")
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func init() {
	index := strings.LastIndex(basepath, "/config")
	if index != -1 {
		ROOT_PATH = strings.Replace(basepath, "/config", "", index)
	}
}

func GetPath(dir string) string {
	return fmt.Sprintf("%s/%s", ROOT_PATH, strings.Trim(dir, "/"))
}
