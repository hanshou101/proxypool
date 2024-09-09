package app

import (
	"log"
	"sync"
	"time"

	"github.com/zu1k/proxypool/internal/cache"
	"github.com/zu1k/proxypool/internal/database"
	"github.com/zu1k/proxypool/pkg/provider"
	"github.com/zu1k/proxypool/pkg/proxy"
)

// 设置时区为中国标准时间
var location, _ = time.LoadLocation("PRC")

// CrawlGo 函数用于抓取代理并进行处理
func CrawlGo() {
	// 创建等待组和代理通道
	wg := &sync.WaitGroup{}
	var pc = make(chan proxy.Proxy)

	// 启动所有getter的抓取任务
	for _, g := range Getters {
		wg.Add(1)
		go g.Get2Chan(pc, wg)
	}

	// 获取现有代理
	proxies := cache.GetProxies("allproxies")
	proxies = append(proxies, database.GetAllProxies()...)

	// 等待所有getter完成并关闭通道
	go func() {
		wg.Wait()
		close(pc)
	}()

	// 从通道接收新的代理
	for node := range pc {
		if node != nil {
			proxies = append(proxies, node)
		}
	}
	// 节点去重
	// 代理去重
	proxies = proxies.Deduplication()
	log.Println("CrawlGo node count:", len(proxies))

	// 清理代理
	proxies = provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.CleanProxies()
	log.Println("CrawlGo cleaned node count:", len(proxies))

	// 对代理进行重命名和排序
	proxies.NameAddCounrty().Sort().NameAddIndex().NameAddTG()
	log.Println("Proxy rename DONE!")

	// 全节点存储到数据库
	// 将所有代理保存到数据库
	database.SaveProxyList(proxies)

	// 更新缓存
	cache.SetProxies("allproxies", proxies)
	cache.AllProxiesCount = proxies.Len()
	log.Println("AllProxiesCount:", cache.AllProxiesCount)

	// 统计各类型代理数量
	cache.SSProxiesCount = proxies.TypeLen("ss")
	log.Println("SSProxiesCount:", cache.SSProxiesCount)
	cache.SSRProxiesCount = proxies.TypeLen("ssr")
	log.Println("SSRProxiesCount:", cache.SSRProxiesCount)
	cache.VmessProxiesCount = proxies.TypeLen("vmess")
	log.Println("VmessProxiesCount:", cache.VmessProxiesCount)
	cache.TrojanProxiesCount = proxies.TypeLen("trojan")
	log.Println("TrojanProxiesCount:", cache.TrojanProxiesCount)

	// 记录最后抓取时间
	cache.LastCrawlTime = time.Now().In(location).Format("2006-01-02 15:04:05")

	// 可用性检测
	// 进行代理可用性检测
	log.Println("Now proceed proxy health check...")
	proxies = proxy.CleanBadProxiesWithGrpool(proxies)
	log.Println("CrawlGo clash usable node count:", len(proxies))

	// 重新索引代理名称
	proxies.NameReIndex()

	// 更新可用代理缓存
	cache.SetProxies("proxies", proxies)
	cache.UsefullProxiesCount = proxies.Len()

	// 生成并缓存Clash格式的代理配置
	cache.SetString("clashproxies", provider.Clash{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide())

	// 生成并缓存Surge格式的代理配置
	cache.SetString("surgeproxies", provider.Surge{
		provider.Base{
			Proxies: &proxies,
		},
	}.Provide())
}
