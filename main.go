// 定义主包
package main

// 导入需要的包
import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof" // 导入 pprof 包,用于性能分析
	"os"

	"github.com/zu1k/proxypool/api"
	"github.com/zu1k/proxypool/internal/app"
	"github.com/zu1k/proxypool/internal/cron"
	"github.com/zu1k/proxypool/internal/database"
	"github.com/zu1k/proxypool/pkg/proxy"
)

// 定义配置文件路径变量
var configFilePath = ""

func main() {
	// 启动一个 goroutine 来运行 pprof 服务器,用于性能分析
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	// 定义命令行参数,用于指定配置文件路径
	flag.StringVar(&configFilePath, "c", "", "path to config file: config.yaml")
	flag.Parse()

	// 如果命令行参数未指定配置文件,则尝试从环境变量获取
	if configFilePath == "" {
		configFilePath = os.Getenv("CONFIG_FILE")
	}
	// 如果环境变量也未指定,则使用默认路径
	if configFilePath == "" {
		configFilePath = "config.yaml"
	}

	// 初始化配置和 Getters
	err := app.InitConfigAndGetters(configFilePath)
	if err != nil {
		// 如果默认路径失败,尝试使用备用路径
		configFilePath = "config/config.yaml"
		err := app.InitConfigAndGetters(configFilePath)
		if err != nil {
			panic(any(err))
		}
	}

	// 初始化数据库表
	database.InitTables()
	// 初始化 GeoIP 数据库
	proxy.InitGeoIpDB()

	fmt.Println("Do the first crawl...")
	// 启动一个 goroutine 执行爬虫任务
	go app.CrawlGo()
	// 启动一个 goroutine 执行定时任务
	go cron.Cron()
	// 运行 API 服务
	api.Run()
}
