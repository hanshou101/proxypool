// 包名声明
package database

// 导入必要的包
import (
	"github.com/zu1k/proxypool/pkg/getter"
	"github.com/zu1k/proxypool/pkg/proxy"
	"gorm.io/gorm"
)

// Proxy 结构体定义，用于数据库模型
type Proxy struct {
	gorm.Model        // 嵌入gorm.Model，提供默认字段（ID, CreatedAt, UpdatedAt, DeletedAt）
	proxy.Base        // 嵌入proxy.Base，包含代理的基本信息
	Link       string // 代理链接
	Identifier string `gorm:"unique"` // 唯一标识符，在数据库中设置为唯一
}

// InitTables 初始化数据库表
func InitTables() {
	// 检查数据库连接是否已建立
	if DB == nil {
		err := connect()
		if err != nil {
			return
		}
	}
	// 自动迁移 Proxy 结构体到数据库，创建或更新表结构
	err := DB.AutoMigrate(&Proxy{})
	if err != nil {
		panic(any(err))
	}
}

// 定义每轮处理的代理数量
const roundSize = 100

// SaveProxyList 保存代理列表到数据库
func SaveProxyList(pl proxy.ProxyList) {
	if DB == nil {
		return
	}

	size := pl.Len()
	// 计算需要多少轮才能处理完所有代理
	round := (size + roundSize - 1) / roundSize

	// 分轮处理代理列表
	for r := 0; r < round; r++ {
		proxies := make([]Proxy, 0, roundSize)
		// 计算本轮要处理的代理范围
		for i, j := r*roundSize, (r+1)*roundSize-1; i < j && i < size; i++ {
			p := pl[i]
			// 创建 Proxy 结构体并添加到切片中
			proxies = append(proxies, Proxy{
				Base:       *p.BaseInfo(),
				Link:       p.Link(),
				Identifier: p.Identifier(),
			})
		}
		// 批量创建代理记录
		DB.Create(&proxies)
	}
}

// GetAllProxies 从数据库获取所有代理
func GetAllProxies() (proxies proxy.ProxyList) {
	proxies = make(proxy.ProxyList, 0)
	if DB == nil {
		return
	}

	// 从数据库中检索所有代理的链接
	proxiesDB := make([]Proxy, 0)
	DB.Select("link").Find(&proxiesDB)

	// 将数据库中的代理转换为 proxy.Proxy 对象
	for _, proxyDB := range proxiesDB {
		if proxiesDB != nil {
			proxies = append(proxies, getter.String2Proxy(proxyDB.Link))
		}
	}
	return
}
