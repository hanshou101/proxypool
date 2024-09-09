// Package cron 提供了定时任务相关的功能
package cron

import (
	"runtime"

	"github.com/jasonlvhit/gocron"
	"github.com/zu1k/proxypool/internal/app"
)

// Cron 函数设置并启动定时任务
func Cron() {
	// 设置每15分钟执行一次crawlTask函数
	_ = gocron.Every(15).Minutes().Do(crawlTask)

	// 启动定时任务并阻塞，等待任务执行
	<-gocron.Start()
}

// crawlTask 函数定义了定时执行的任务
func crawlTask() {
	// 初始化配置和获取器
	_ = app.InitConfigAndGetters("")

	// 执行爬虫任务
	app.CrawlGo()

	// 清空获取器，释放资源
	app.Getters = nil

	// 手动触发垃圾回收，释放内存
	runtime.GC()
}
