package options

import (
	"fmt"
	"github.com/caoyingjunz/pixiulib/config"
	"github.com/caoyingjunz/rainbow/pkg/controller"
	"github.com/caoyingjunz/rainbow/pkg/controller/image"
	rainbowdb "github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"os"
)

const (
	defaultConfigFile = "/etc/rainbow/config.yaml"

	maxIdleConns = 10
	maxOpenConns = 100
)

type Options struct {
	ComponentConfig image.Config

	HttpEngine *gin.Engine

	db      *gorm.DB
	Factory rainbowdb.ShareDaoFactory

	Controller controller.RainbowInterface

	ConfigFile string
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

	// 注册依赖组件
	if err := o.register(); err != nil {
		return err
	}

	o.Controller = controller.New("test", o.Factory)
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
