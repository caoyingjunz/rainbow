package options

import (
	"fmt"
	"os"

	"github.com/caoyingjunz/pixiulib/config"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/controller"
	rainbowdb "github.com/caoyingjunz/rainbow/pkg/db"
)

const (
	defaultConfigFile = "/etc/rainbow/config.yaml"
	defaultDataDir    = "/data"
	defaultListen     = 8090

	maxIdleConns = 10
	maxOpenConns = 100
)

type Options struct {
	ComponentConfig rainbowconfig.Config
	ConfigFile      string

	db      *gorm.DB
	Factory rainbowdb.ShareDaoFactory

	HttpEngine *gin.Engine
	Controller controller.RainbowInterface
}

func NewOptions(configFile string) (*Options, error) {
	return &Options{
		HttpEngine: gin.Default(),
		ConfigFile: configFile,
	}, nil
}

// Complete completes all the required options
func (o *Options) Complete() error {
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

	if len(o.ComponentConfig.Agent.DataDir) == 0 {
		o.ComponentConfig.Agent.DataDir = defaultDataDir
	}
	if o.ComponentConfig.Default.Listen == 0 {
		o.ComponentConfig.Default.Listen = defaultListen
	}

	// 注册依赖组件
	if err := o.register(); err != nil {
		return err
	}

	o.Controller = controller.New(o.ComponentConfig, o.Factory)
	return nil
}

func (o *Options) register() error {
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
