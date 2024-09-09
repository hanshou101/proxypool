package api

import (
	// 导入必要的包
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/zu1k/proxypool/config"
	binhtml "github.com/zu1k/proxypool/internal/bindata/html"
	"github.com/zu1k/proxypool/internal/cache"
	"github.com/zu1k/proxypool/pkg/provider"
)

// 定义版本号
const version = "v0.3.8"

// 定义全局路由变量
var router *gin.Engine

// setupRouter 函数用于设置和配置路由
func setupRouter() {
	// 设置gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	// 创建一个新的gin引擎
	router = gin.New()
	// 使用gin的恢复中间件
	router.Use(gin.Recovery())

	// 加载HTML模板
	temp, err := loadTemplate()
	if err != nil {
		panic(any(err))
	}
	router.SetHTMLTemplate(temp)

	// 设置首页路由
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/index.html", gin.H{
			// 传递各种数据到模板
			"domain":               config.Config.Domain,
			"getters_count":        cache.GettersCount,
			"all_proxies_count":    cache.AllProxiesCount,
			"ss_proxies_count":     cache.SSProxiesCount,
			"ssr_proxies_count":    cache.SSRProxiesCount,
			"vmess_proxies_count":  cache.VmessProxiesCount,
			"trojan_proxies_count": cache.TrojanProxiesCount,
			"useful_proxies_count": cache.UsefullProxiesCount,
			"last_crawl_time":      cache.LastCrawlTime,
			"version":              version,
		})
	})

	// 设置Clash页面路由
	router.GET("/clash", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	// 设置Surge页面路由
	router.GET("/surge", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	// 设置Clash配置路由
	router.GET("/clash/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash-config.yaml", gin.H{
			"domain": config.Config.Domain,
		})
	})

	// 设置Surge配置路由
	router.GET("/surge/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.conf", gin.H{
			"domain": config.Config.Domain,
		})
	})

	// 设置Clash代理路由
	router.GET("/clash/proxies", func(c *gin.Context) {
		// 获取查询参数
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		text := ""

		// 根据查询参数返回不同的代理列表
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" {
			// 如果没有指定参数，返回缓存的所有代理
			text = cache.GetString("clashproxies")
			if text == "" {
				proxies := cache.GetProxies("proxies")
				clash := provider.Clash{
					provider.Base{
						Proxies: &proxies,
					},
				}
				text = clash.Provide()
				cache.SetString("clashproxies", text)
			}
		} else if proxyTypes == "all" {
			// 如果type为all，返回所有类型的代理
			proxies := cache.GetProxies("allproxies")
			clash := provider.Clash{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = clash.Provide()
		} else {
			// 否则根据指定的参数筛选代理
			proxies := cache.GetProxies("proxies")
			clash := provider.Clash{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = clash.Provide()
		}
		c.String(200, text)
	})

	// 设置Surge代理路由，逻辑同Clash
	router.GET("/surge/proxies", func(c *gin.Context) {
		// ... (类似于Clash代理的逻辑)
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" {
			text = cache.GetString("surgeproxies")
			if text == "" {
				proxies := cache.GetProxies("proxies")
				surge := provider.Surge{
					provider.Base{
						Proxies: &proxies,
					},
				}
				text = surge.Provide()
				cache.SetString("surgeproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := cache.GetProxies("allproxies")
			surge := provider.Surge{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = surge.Provide()
		} else {
			proxies := cache.GetProxies("proxies")
			surge := provider.Surge{
				provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
				},
			}
			text = surge.Provide()
		}
		c.String(200, text)
	})

	// 设置SS订阅路由
	router.GET("/ss/sub", func(c *gin.Context) {
		proxies := cache.GetProxies("proxies")
		ssSub := provider.SSSub{
			provider.Base{
				Proxies: &proxies,
				Types:   "ss",
			},
		}
		c.String(200, ssSub.Provide())
	})

	// 设置SSR订阅路由
	router.GET("/ssr/sub", func(c *gin.Context) {
		// ... (类似于SS订阅的逻辑)
		proxies := cache.GetProxies("proxies")
		ssrSub := provider.SSRSub{
			provider.Base{
				Proxies: &proxies,
				Types:   "ssr",
			},
		}
		c.String(200, ssrSub.Provide())
	})

	// 设置Vmess订阅路由
	router.GET("/vmess/sub", func(c *gin.Context) {
		// ... (类似于SS订阅的逻辑)
		proxies := cache.GetProxies("proxies")
		vmessSub := provider.VmessSub{
			provider.Base{
				Proxies: &proxies,
				Types:   "vmess",
			},
		}
		c.String(200, vmessSub.Provide())
	})

	// 设置获取单个代理链接的路由
	router.GET("/link/:id", func(c *gin.Context) {
		idx := c.Param("id")
		proxies := cache.GetProxies("allproxies")
		id, err := strconv.Atoi(idx)
		if err != nil {
			c.String(500, err.Error())
		}
		if id >= proxies.Len() {
			c.String(500, "id too big")
		}
		c.String(200, proxies[id].Link())
	})
}

// Run 函数用于启动HTTP服务
func Run() {
	setupRouter()
	// 获取环境变量中的端口，如果没有则使用默认端口8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}

// loadTemplate 函数用于加载HTML模板
func loadTemplate() (t *template.Template, err error) {
	// 恢复assets/html目录下的所有资产
	_ = binhtml.RestoreAssets("", "assets/html")
	t = template.New("")
	// 遍历所有资产并解析为模板
	for _, fileName := range binhtml.AssetNames() {
		data := binhtml.MustAsset(fileName)
		t, err = t.New(fileName).Parse(string(data))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
