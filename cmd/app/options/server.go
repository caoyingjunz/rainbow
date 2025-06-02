package options

import (
	"fmt"
	"os"

	"github.com/caoyingjunz/pixiulib/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/controller"
	rainbowdb "github.com/caoyingjunz/rainbow/pkg/db"
)

type ServerOptions struct {
	ComponentConfig rainbowconfig.Config
	ConfigFile      string

	db      *gorm.DB
	Factory rainbowdb.ShareDaoFactory

	RedisClient *redis.Client

	HttpEngine *gin.Engine
	Controller controller.RainbowInterface
}

func NewServerOptions(configFile string) (*ServerOptions, error) {
	return &ServerOptions{
		HttpEngine: gin.Default(),
		ConfigFile: configFile,
	}, nil
}

// Complete completes all the required options
func (o *ServerOptions) Complete() error {
	// 配置文件优先级: 默认配置，环境变量，命令行
	if len(o.ConfigFile) == 0 {
		// Try to read config file path from env.
		if cfgFile := os.Getenv("ConfigFile"); cfgFile != "" {
			o.ConfigFile = cfgFile
		} else {
			o.ConfigFile = defaultConfigFile
		}
	}

	c := config.New()
	c.SetConfigFile(o.ConfigFile)
	c.SetConfigType("yaml")

	if err := c.Binding(&o.ComponentConfig); err != nil {
		klog.Fatal(err)
	}

	// 注册依赖组件
	if err := o.register(); err != nil {
		return err
	}

	o.Controller = controller.New(o.ComponentConfig, o.Factory, o.RedisClient)
	return nil
}

func (o *ServerOptions) register() error {
	sqlConfig := o.ComponentConfig.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		sqlConfig.User,
		sqlConfig.Password,
		sqlConfig.Host,
		sqlConfig.Port,
		sqlConfig.Name)

	opt := &gorm.Config{}
	db, err := gorm.Open(mysql.Open(dsn), opt)
	if err != nil {
		return err
	}
	o.db = db

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)

	o.Factory, err = rainbowdb.NewDaoFactory(db, true)
	return err
}

func (o *ServerOptions) registerRedis() error {
	redisConfig := o.ComponentConfig.Redis
	o.RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.Db,
	})

	return nil
}
