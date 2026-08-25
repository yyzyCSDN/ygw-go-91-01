基于 Go 实现的机场供油管网监控系统项目，一款机场供油控制服务，完成加油栓压力/流量监测、泵组控制、油品切换与泄漏联锁管理。

# FuelHydrantMonitor

机场供油管网监控服务。启动后监听 HTTP 端口，提供供油状态、泵组、加油栓、油品切换、泄漏联锁、告警与运行记录的查询与操作接口，并提供 web/console.html 控制台页面。

## 构建与运行

```bash
go build -mod=vendor ./...
go run -mod=vendor ./cmd/fuelhydrant -addr 0.0.0.0:8091 -dir ./data
```

启动后访问 http://127.0.0.1:8091/ 查看控制台。
