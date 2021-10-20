package service

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/yaoapp/gou"
	"github.com/yaoapp/xiang/chart"
	"github.com/yaoapp/xiang/config"
	"github.com/yaoapp/xiang/share"
	"github.com/yaoapp/xiang/table"
)

// Watch 监听应用目录文件变更
func Watch(cfg config.Config) {
	WatchEngine(cfg.Path)
	WatchModel(cfg.RootModel, "")
	WatchAPI(cfg.RootAPI, "")
	WatchFlow(cfg.RootFLow, "")
	WatchPlugin(cfg.RootPlugin)
	WatchTable(cfg.RootTable, "")
	WatchChart(cfg.RootChart, "")
}

// WatchEngine 监听监听引擎内建数据变更
func WatchEngine(root string) {
	if !strings.HasPrefix(root, "fs://") && strings.Contains(root, "://") {
		return
	}
	root = strings.TrimPrefix(root, "fs://")
	root = share.DirAbs(root)

	WatchModel(filepath.Join(root, "models"), "xiang.")
	WatchAPI(filepath.Join(root, "apis"), "xiang.")
	WatchFlow(filepath.Join(root, "flows"), "xiang.")
	WatchTable(filepath.Join(root, "tables"), "xiang.")
}

// WatchModel 监听业务接口更新
func WatchModel(root string, prefix string) {
	if share.DirNotExists(root) {
		return
	}
	root = share.DirAbs(root)
	go share.Watch(root, func(op string, filename string) {

		if !strings.HasSuffix(filename, ".json") {
			return
		}
		if op == "write" || op == "create" {
			name := prefix + share.SpecName(root, filename)
			content := share.ReadFile(filename)
			gou.LoadModel(string(content), name) // Reload
			log.Printf("Model %s 已重新加载完毕", name)

		} else if op == "remove" || op == "rename" {
			name := prefix + share.SpecName(root, filename)
			if _, has := gou.Models[name]; has {
				delete(gou.Models, name)
				log.Printf("Model %s 已经移除", name)
			}
		}
	})
}

// WatchAPI 监听业务接口更新
func WatchAPI(root string, prefix string) {
	if share.DirNotExists(root) {
		return
	}
	root = share.DirAbs(root)
	go share.Watch(root, func(op string, filename string) {

		if !strings.HasSuffix(filename, ".json") {
			return
		}

		if op == "write" || op == "create" {
			name := prefix + share.SpecName(root, filename)
			content := share.ReadFile(filename)
			gou.LoadAPI(string(content), name) // Reload
			log.Printf("API %s 已重新加载完毕", name)

		} else if op == "remove" || op == "rename" {
			name := prefix + share.SpecName(root, filename)
			if _, has := gou.APIs[name]; has {
				delete(gou.APIs, name)
				log.Printf("API %s 已经移除", name)
			}
		}

		// 重启服务器
		if op == "write" || op == "create" || op == "remove" || op == "rename" {
			Stop(func() {
				log.Printf("服务器重启完毕")
				go Start()
			})
		}
	})
}

// WatchFlow 监听业务逻辑变更
func WatchFlow(root string, prefix string) {
	if share.DirNotExists(root) {
		return
	}
	root = share.DirAbs(root)
	go share.Watch(root, func(op string, filename string) {
		if !strings.HasSuffix(filename, ".json") && !strings.HasSuffix(filename, ".js") {
			return
		}

		if strings.HasSuffix(filename, ".js") {
			name := prefix + share.SpecName(root, filename)
			filename = name + ".flow.json"
		}

		if op == "write" || op == "create" {
			name := prefix + share.SpecName(root, filename)
			content := share.ReadFile(filename)
			gou.LoadFlow(string(content), name) // Reload
			log.Printf("Flow %s 已重新加载完毕", name)

		} else if op == "remove" || op == "rename" {
			name := prefix + share.SpecName(root, filename)
			if _, has := gou.Flows[name]; has {
				delete(gou.Flows, name)
				log.Printf("Flow %s 已经移除", name)
			}
		}
	})
}

// WatchPlugin 监听业务插件变更
func WatchPlugin(root string) {
	if share.DirNotExists(root) {
		return
	}
	root = share.DirAbs(root)
	go share.Watch(root, func(op string, filename string) {

		if !strings.HasSuffix(filename, ".so") {
			return
		}

		if op == "write" || op == "create" {
			name := share.SpecName(root, filename)
			gou.LoadPlugin(filename, name) // Reload
			log.Printf("Plugin %s 已重新加载完毕", name)

		} else if op == "remove" || op == "rename" {
			name := share.SpecName(root, filename)
			if _, has := gou.Plugins[name]; has {
				delete(gou.Plugins, name)
				log.Printf("Plugin %s 已经移除", name)
			}
		}
	})
}

// WatchTable 监听数据表格更新
func WatchTable(root string, prefix string) {
	if share.DirNotExists(root) {
		return
	}
	root = share.DirAbs(root)
	go share.Watch(root, func(op string, filename string) {

		if !strings.HasSuffix(filename, ".json") {
			return
		}

		if op == "write" || op == "create" {
			name := prefix + share.SpecName(root, filename)
			content := share.ReadFile(filename)
			table.LoadTable(string(content), name) // Reload Table

			api, has := gou.APIs["xiang.table"]
			if has {
				api.Reload() // 重载API
			}
			log.Printf("数据表格 %s 已重新加载完毕", name)

		} else if op == "remove" || op == "rename" {
			name := prefix + share.SpecName(root, filename)
			if _, has := table.Tables[name]; has {
				delete(table.Tables, name)
				log.Printf("数据表格 %s 已经移除", name)
			}
		}

		// 重启服务器
		if op == "write" || op == "create" || op == "remove" || op == "rename" {
			Stop(func() {
				log.Printf("服务器重启完毕")
				go Start()
			})
		}
	})
}

// WatchChart 监听分析图表更新
func WatchChart(root string, prefix string) {
	if share.DirNotExists(root) {
		return
	}
	root = share.DirAbs(root)
	go share.Watch(root, func(op string, filename string) {
		if !strings.HasSuffix(filename, ".json") && !strings.HasSuffix(filename, ".js") {
			return
		}

		if strings.HasSuffix(filename, ".js") {
			name := prefix + share.SpecName(root, filename)
			filename = name + ".chart.json"
		}

		if op == "write" || op == "create" {
			name := prefix + share.SpecName(root, filename)
			content := share.ReadFile(filename)
			chart.LoadChart(content, name) // Relaod

			api, has := gou.APIs["xiang.chart"]
			if has {
				fmt.Println("Reload API:", "--")
				api.Reload() // 重载API
			}
			log.Printf("Chart %s 已重新加载完毕", name)

		} else if op == "remove" || op == "rename" {
			name := prefix + share.SpecName(root, filename)
			if _, has := chart.Charts[name]; has {
				delete(chart.Charts, name)
				log.Printf("Chart %s 已经移除", name)
			}
		}

		// 重启服务器
		if op == "write" || op == "create" || op == "remove" || op == "rename" {
			Stop(func() {
				log.Printf("服务器重启完毕")
				go Start()
			})
		}
	})
}