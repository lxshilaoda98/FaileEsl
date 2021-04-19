package database

import (
	"database/sql"
	"fmt"
	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)
import _ "github.com/denisenkom/go-mssqldb"

var SqlDB *sql.DB

var DBDriver = "mysql" //<mysql/mssql>

type AppConfig struct {
	DataBase string
}

func GetIVRConfig() (config *viper.Viper) {
	config = viper.New()
	config.AddConfigPath("./")
	config.SetConfigName("config")
	config.SetConfigType("json")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := config.ReadInConfig(); err != nil {
			panic(err)
		}
	})
	//直接反序列化为Struct
	var configjson AppConfig
	if err := config.Unmarshal(&configjson); err != nil {
		fmt.Println(err)
	}
	return
}
func init() {
	config := GetIVRConfig()
	connString := ""
	if DBDriver == "mssql" {
		connString = "server=127.0.0.1;port=1433;database=TF_LFT_CMS;user id=sa;password=sa@01;encrypt=disable"
	} else if DBDriver == "mysql" {
		connString = config.GetString("AppConfig.DataBase")
	}
	var err error
	SqlDB, err = sql.Open(DBDriver, connString)
	if err != nil {
		fmt.Println("数据库连接异常:err..>", err.Error())
	}
	err = SqlDB.Ping()
	if err != nil {
		fmt.Println("[ping]数据库连接异常:err..>", err.Error())
	}
}
