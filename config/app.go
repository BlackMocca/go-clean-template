package config

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Viper    *viper.Viper
	PsqlConn *sql.DB
}

func NewConfig() *Config {
	v := viper.New()
	v.SetConfigName(".env")
	v.AddConfigPath(".")
	loadEnvConfig(v)

	/* Loading Other Config */

	return &Config{
		Viper: v,
	}
}

func loadEnvConfig(v *viper.Viper) {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		panic(".env file does not exist")
	}

	file, _ := os.Open(".env")
	err := v.ReadConfig(file)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s \n", err))
	}

}

func (c *Config) GetEnv(key string, defaultValue string) string {
	if has := c.Viper.IsSet(key); has {
		return c.Viper.GetString(key)
	}
	return defaultValue
}