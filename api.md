# JetGeo Server API 文档

版本: 1.2.2-SNAPSHOT  
最后更新: 2025-09-26

## 概述
JetGeo Server 提供经纬度反向解析(Reverse Geocoding)能力, 当前仅提供一个 HTTP 接口: 通过纬度(lat)与经度(lng)获取行政区划信息 (省/市/区县)。

后续可扩展: 增加批量接口、街道级定位、健康检查等。

---
## 基础信息
- 基础 URL: `http://<host>:<port>` (默认端口 `8080`)
- 接口前缀: 无额外全局前缀
- 数据格式: `application/json; charset=UTF-8`
- 鉴权: 当前无需鉴权

---
## 配置 (application.yml)
```yaml
server:
  port: 8080

jetgeo:
  geo-data-parent-path: ${GEO_DATA_PATH:./data/geodata}  # 数据目录, 环境变量 GEO_DATA_PATH 优先
  level: district  # 解析定位的最细层级: province / city / district
```
可用 JVM 启动覆盖:
```
--server.port=9090 --jetgeo.level=city
```
或环境变量指定数据目录:
```
GEO_DATA_PATH=D:/s2/jetgeo/data/geodata
```

---
## 1. 反向地理编码接口
### 请求
- 方法: `GET`
- 路径: `/api/reverse`
- 描述: 根据纬度与经度返回行政区划信息。

### 请求参数
| 名称 | 位置 | 类型 | 必填 | 示例 | 说明 |
|------|------|------|------|------|------|
| lat  | query | double | 是 | 39.9042 | 纬度 (北纬为正) |
| lng  | query | double | 是 | 116.4074 | 经度 (东经为正) |

### 成功响应 (HTTP 200)
响应体对应模型 `GeoInfo`:
| 字段 | 类型 | 示例 | 说明 |
|------|------|------|------|
| formatAddress | string | 北京市东城区 | 组合格式化后的地址 (视实现可能为空) |
| province | string | 北京市 | 省/直辖市/自治区名称 |
| provinceCode | string | 110000 | 省级行政区代码 |
| city | string | 北京市 | 地级市名称 (直辖市可能与 province 相同) |
| cityCode | string | 110100 | 地级市行政代码 |
| district | string | 东城区 | 区/县名称 |
| districtCode | string | 110101 | 区/县行政代码 |
| street | string | (空) | 预留字段, 当前未实现街道级定位 |
| streetCode | string | (空) | 预留字段, 当前未实现街道级编码 |
| adcode | string | 110101 | 实际匹配到的最小级别行政区代码 |
| level | string(enum) | district | 实际匹配到的最细层级: province/city/district |

> 注意: `level` 表示返回数据达到的真实层级。如果请求配置为 `district` 但该点无法匹配到区县, 可能退化为 `city` 或 `province`。

### 失败响应
| HTTP 状态码 | 说明 | 返回体 |
|-------------|------|--------|
| 404 | 未找到匹配的行政区划 (坐标超出数据覆盖范围或数据缺失) | 无响应体 |
| 400 | 参数格式错误 (例如不能解析为 double) | Spring 默认错误 JSON |
| 500 | 服务器内部错误 | Spring 默认错误 JSON |

### 示例
请求:
```
GET /api/reverse?lat=39.9042&lng=116.4074
Host: localhost:8080
```
cURL:
```bash
curl "http://localhost:8080/api/reverse?lat=39.9042&lng=116.4074"
```
可能的响应:
```json
{
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
  "level": "district"
}
```

区县未命中时示例 (退化到市级):
```json
{
  "formatAddress": "某省某市",
  "province": "某省",
  "provinceCode": "340000",
  "city": "某市",
  "cityCode": "340100",
  "district": "",
  "districtCode": "",
  "street": "",
  "streetCode": "",
  "adcode": "340100",
  "level": "city"
}
```

### 错误示例
```
GET /api/reverse?lat=abc&lng=116
```
返回 (400):
```json
{
  "timestamp": "2025-09-26T11:15:22.123+00:00",
  "status": 400,
  "error": "Bad Request",
  "message": "Failed to convert value of type 'java.lang.String' to required type 'double'; ...",
  "path": "/api/reverse"
}
```

---
## 部署 / 运行指引简述
```powershell
# 设置数据目录 (例如已经解压 data/geodata)
$env:GEO_DATA_PATH = "D:\s2\jetgeo\data\geodata"
# 运行
java -jar .\jetgeo-server\target\jetgeo-server-1.2.2-SNAPSHOT.jar --server.port=8080
```

---
## 版本演进建议 (非实现)
- 支持批量坐标 POST JSON
- 增加健康检查 `/actuator/health`
- 增加街道级数据与缓存
- 增加返回坐标匹配精度 / 数据源版本号

---
## 许可证
见根目录 `LICENSE`。

---
如需补充或生成 OpenAPI/Swagger 文档, 可在 `jetgeo-server` 引入 `springdoc-openapi` 依赖, 再告知我协助配置。
