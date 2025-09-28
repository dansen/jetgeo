# jetgeo-go

Go 版本的 JetGeo 反向地理编码引擎 (初始骨架)。

## 功能
- 省级边界预加载, 市/区延迟缓存加载
- S2 多边形包含判断 (使用 github.com/golang/geo/s2)
- Gin HTTP 服务 `/api/reverse?lat=..&lng=..`
- 统一 JSON 响应格式: `{"code":0,"message":"OK","data":{...}}`
- 请求详细日志 (method/path/status/duration/size/remote) + 反向解析命中/未命中日志

## 快速运行
```
# 假设已经存在数据目录 ./data/geodata (province/city/district)
go mod tidy

go run ./cmd --data ./data/geodata --level district --addr :8080

curl "http://localhost:8080/api/reverse?lat=39.9042&lng=116.4074"
```

## 环境变量
- GEO_DATA_PATH: 数据根目录 (与 --data 二选一, flag 优先)
- JETGEO_LEVEL: province|city|district
 - LOG_LEVEL: debug|info (默认为 info, debug 将输出每次调用的坐标与结果详情)

## 日志
默认使用 zap production 配置:
- 访问日志: 统一格式字段: method, path, status, size, duration, remote
- 业务日志:
	- reverse_request (debug) 输入原始参数
	- reverse_hit (info) 命中行政区 (包含 province/city/district/adcode/level)
	- reverse_not_found (info) 未命中
	- reverse_missing_param / reverse_parse_error (warn)

开启 debug:
```
set LOG_LEVEL=debug  # Windows PowerShell: $env:LOG_LEVEL="debug"
```

## 响应格式
成功示例:
```json
{
	"code": 0,
	"message": "OK",
	"data": {
		"formatAddress": "北京市东城区",
		"province": "北京市",
		"provinceCode": "110000",
		"city": "北京市",
		"cityCode": "110100",
		"district": "东城区",
		"districtCode": "110101",
		"street": "",
		"streetCode": "",
		"adcode": "110101",
		"level": 2
	}
}
```

错误示例 (缺少参数):
```json
{"code":1001,"message":"missing lat or lng"}
```

未命中示例:
```json
{"code":1404,"message":"not found"}
```

## 后续计划
- 增加 metrics / pprof
- 增加缓存刷新 goroutine / 指标
- 增加测试与 benchmark
- 增加可配置的访问日志字段与采样

---
此仓库内容为从 Java 版本迁移的初稿设计与骨架, 具体算法行为须通过测试验证。
