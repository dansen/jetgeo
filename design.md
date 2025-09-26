# jetgeo-go 设计文档 (初稿)

日期: 2025-09-26
目标: 将现有 Java 版本 `jetgeo` (反向地理编码, S2 多边形包含判断 + 分级缓存加载) 迁移/重写为 Go 版本, 保持功能等价, 在性能/内存/部署形态上更轻量。

---
## 1. Java 版本功能要点回顾
| 模块 | 作用 | 关键点 |
|------|------|--------|
| jetgeo-core | 核心算法与数据载入 | 省级全量预加载, 市/区延迟加载 + Guava LoadingCache, S2 多边形 contains, JSON 多边形文件解析 |
| jetgeo-spring-boot-starter | Spring Boot 自动装配 | 绑定 `jetgeo.*` 配置并注入 `JetGeo` Bean |
| jetgeo-server | REST 接口 | `/api/reverse?lat=..&lng=..` 返回 `GeoInfo` |
| 数据格式 | `level_regionCode[_regionName].json` | 省: `province_110000_北京市.json`; 市: `city_110000_110100_北京市.json`; 区: `district_110100_110101_东城区.json` |

核心类/职责:
- JetGeo: 初始化+持有缓存, 提供 `GetGeoInfo(lat,lng)`
- RegionCache / RegionCacheLoader: 单区域多 polygon S2Region 集合; 通过父 code 延迟加载子级文件
- Utils: 解析文件名 -> 行政区信息, JSON -> 多边形点列 -> S2Region
- S2Util: 包装 S2 polygon/cell contains
- GeoInfo: 结果聚合 (省/市/区 & adcode & level)

---
## 2. Go 版本总体设计
### 2.1 设计目标
- 模块化: 将核心算法与接口解耦, 支持作为库引用或启动独立 HTTP 服务
- 高性能: 省级一次性加载; 子级采用按需 + LRU/TTL 缓存
- 可观测: 基础 metrics / pprof / 可选日志等级
- 易配置: 支持 ENV > CLI Flag > YAML(可选) 的优先级
- 并发安全: 读多写少场景, 使用 sync.RWMutex 或 sync.Map + 封装

### 2.2 目录结构 (建议)
```
jetgeo-go/
  cmd/
    jetgeo-server/
      main.go               # 解析配置 & 启动 HTTP
  internal/
    geo/
      jetgeo.go             # JetGeo 主结构体与构造
      cache.go              # 缓存接口与实现 (省/市/区)
      region.go             # RegionCache 定义与加载
      loader.go             # 基于父 code 延迟加载逻辑
      s2util.go             # S2 封装
      geomjson.go           # JSON 多边形读取与解析
      model.go              # GeoInfo, Level 枚举, RegionInfo 等
      config.go             # Config 结构体与默认值/加载顺序
    httpx/
      server.go             # HTTP Server 封装 (路由, 中间件)
      handler_reverse.go    # /api/reverse 处理
      response.go           # 统一响应/错误
      middleware.go         # 日志/恢复/metrics
    util/
      logger.go             # 日志初始化
      envflag.go            # flag+env 合并
  pkg/
    jetgeo/                 # 对外公开 API (薄包装, import friendly)
      api.go
  data/
    (可放示例数据 subset 用于测试)
  testdata/
    polygons/...
  docs/
    api.md                  # (从 Java 复制/裁剪)
    design.md               # 设计文档 (本文件复制一份)
  scripts/
    build.sh
    run.sh
  go.mod
  go.sum
  Makefile (可选)
  README.md
```

说明:
- internal/geo 下是核心实现, 对外不直接暴露; pkg/jetgeo 暴露 `type JetGeo interface { Reverse(lat,lng float64) (*GeoInfo, bool) }` 方便第三方使用
- testdata 存放小规模数据, 避免全量政区数据拖慢 CI

---
## 3. Go 模型与 Java 映射
| Java | Go | 说明 |
|------|----|------|
| LevelEnum | `type Level int` + const (Province/City/District) | iota 定义, String() 方法 |
| GeoInfo | `type GeoInfo struct { ... }` | 字段名导出; JSON tag 驼峰保持一致 |
| RegionCache | `type RegionCache struct { Name ParentCode Code Level Polygons []S2RegionLike }` | 使用 `[]*s2.Polygon` 或自定义包装 |
| RegionInfo | 合入 RegionCache 加载过程 | 单次解析即可 |
| JetGeo | `type Engine struct { cfg Config province map[string]*RegionCache cityCache *LoadingCache districtCache *LoadingCache ... }` | Engine.Reverse() |
| RegionCacheLoader | 函数: `LoadChildren(level Level, parent string) ([]*RegionCache, error)` | 加载逻辑独立 |
| Utils.jsonPointLists2Tuple2Lists | 直接解析成 `[][]LatLng` | 简化 |
| S2Util | `s2util.go` + 第三方库 (github.com/golang/geo/s2) | contains 判断优化 |

### 3.1 枚举
```go
type Level int
const (
    LevelProvince Level = iota
    LevelCity
    LevelDistrict
)
func (l Level) String() string { ... }
```

### 3.2 GeoInfo 结构
```go
type GeoInfo struct {
    FormatAddress string `json:"formatAddress"`
    Province      string `json:"province"`
    ProvinceCode  string `json:"provinceCode"`
    City          string `json:"city"`
    CityCode      string `json:"cityCode"`
    District      string `json:"district"`
    DistrictCode  string `json:"districtCode"`
    Street        string `json:"street"`
    StreetCode    string `json:"streetCode"`
    Adcode        string `json:"adcode"`
    Level         Level  `json:"level"`
}
```

### 3.3 RegionCache
```go
type RegionCache struct {
    Name       string
    ParentCode string
    Code       string
    Level      Level
    Polygons   []*s2.Polygon // 或包装类型
}
func (r *RegionCache) Contains(lat, lng float64) bool { ... }
func (r *RegionCache) CacheKey() string { ... }
```

---
## 4. 缓存策略
Java 使用 Guava LoadingCache (带 expireAfterAccess + refreshAfterWrite)。Go 方案:
1. 使用自实现 + sync.Map + 元数据(上次访问时间) + 后台清理 goroutine
2. 或使用第三方: `github.com/patrickmn/go-cache` / `github.com/dgraph-io/ristretto` / `github.com/hashicorp/golang-lru`

推荐: 自定义轻量 LRU + TTL (子级数据相对有限) 逻辑:
```go
type entry struct {
    regions []*RegionCache
    loadedAt time.Time
}

type LoadingCache struct {
    mu sync.RWMutex
    data map[string]*entry
    ttl time.Duration
    loader func(key string) ([]*RegionCache, error)
}

func (c *LoadingCache) Get(key string) ([]*RegionCache, error) {
    c.mu.RLock(); e, ok := c.data[key]; c.mu.RUnlock()
    if ok && time.Since(e.loadedAt) < c.ttl { return e.regions, nil }
    // load / reload
}
```
刷新策略: 惰性 + 后台定期 scan (interval = ttl/2)。

---
## 5. S2 集成
使用官方 Go 包: `github.com/golang/geo/s2`
多边形构建步骤:
1. 解析 JSON => `[][]LatLng`
2. 每个环构造 `s2.LoopFromPoints([]s2.Point)` (需要经纬度 -> s2.LatLng -> s2.Point)
3. 单环区域 => `s2.PolygonFromLoops([]*s2.Loop{loop})`
4. Contains: `poly.ContainsPoint(s2.LatLngFromDegrees(lat,lng).Normalized().Point())`

注意: Java 版本假设多边形点序为逆时针 (S2Loop 方向)。Go 需要校正: 若 `loop.IsHole()` 需 `loop.Invert()`. 可写校验函数。

---
## 6. JSON 数据文件解析
文件命名规则解析：`{level}_{parent?}_{code}_{name}.json`
- province: province_<code>_<name>.json
- city: city_<provinceCode>_<code>_<name>.json
- district: district_<cityCode>_<code>_<name>.json

Go 解析逻辑伪代码:
```go
func ParseFileName(path string) (info RegionMeta, err error) {
    base := filepath.Base(path)
    name := strings.TrimSuffix(base, filepath.Ext(base))
    parts := strings.Split(name, "_")
    switch parts[0] { ... }
}
```

JSON 内容: `[[{"lat": .., "lng": ..}, ...], [...]]` -> 多个环。
结构:
```go
type jsonPoint struct { Lat float64 `json:"lat"`; Lng float64 `json:"lng"` }
```
解析: `var polys [][]jsonPoint`

---
## 7. Engine (核心)
```go
type Config struct {
    GeoDataPath string
    Level       Level
    CityTTL     time.Duration
    DistrictTTL time.Duration
    // S2 tuning (可选)
}

type Engine struct {
    cfg Config
    province map[string]*RegionCache
    cityCache     *LoadingCache
    districtCache *LoadingCache
}

func NewEngine(cfg Config) (*Engine, error) {
    // 1. load provinces
    // 2. init caches with loaders
}

func (e *Engine) Reverse(lat,lng float64) (*GeoInfo, bool) {
    // 1. 遍历省 map
    // 2. 命中继续 city
    // 3. 命中继续 district
    // 4. 构造 GeoInfo
}
```
优化点: 省数量有限 (≈34) 顺序扫描即可; 可后期加 S2 cell 索引 (map[cellTokenPrefix] -> candidate provinces)。初版不做提前优化。

---
## 8. HTTP 层设计
使用 `net/http` + `chi` (或 gin)。建议 chi (轻量)。
路由:
```
GET /api/reverse?lat=..&lng=..
GET /healthz
GET /metrics (可选 / Prometheus)
```
示例 handler:
```go
func reverseHandler(e *Engine) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
     latStr := r.URL.Query().Get("lat")
     lngStr := r.URL.Query().Get("lng")
     // parse -> call e.Reverse -> write JSON / 404
  }
}
```

启动流程 main.go:
```go
func main() {
  cfg := LoadConfig()
  eng, err := NewEngine(cfg)
  // create router, register, graceful shutdown
}
```

---
## 9. 配置加载顺序
1. 环境变量 (例如 GEO_DATA_PATH, JETGEO_LEVEL)
2. 命令行 flag (覆盖 env)
3. 可选 YAML (若提供 --config)
4. 默认值

---
## 10. 错误与日志
- 使用 `errors.Wrap` (或 Go 1.20 errors.Join) 保留上下文
- 日志库: 标准库 log 或 `zap` (开发/生产 JSON 可选)
- 全局日志等级通过 ENV: `LOG_LEVEL=info|debug`

---
## 11. 测试策略
| 测试类型 | 内容 |
|----------|------|
| 单元 | 文件名解析, JSON 解析, Polygon contains, 缓存刷新/过期 |
| 集成 | Engine 初始化 + 反向查询多坐标, 404 场景 |
| 基准 | Reverse 1000 点 (命中/未命中), 缓存冷/热对比 |
| 竞态 | -race 模式下并发查询 |

示例基准:
```go
func BenchmarkReverseHot(b *testing.B) {
  for i:=0; i<b.N; i++ { eng.Reverse(39.9,116.4) }
}
```

---
## 12. 初始文件清单 (草案)
| 文件 | 说明 |
|------|------|
| cmd/jetgeo-server/main.go | 程序入口 |
| internal/geo/config.go | Config 定义+加载 |
| internal/geo/level.go | Level 枚举 & String |
| internal/geo/model.go | GeoInfo 等结构体 |
| internal/geo/region.go | RegionCache 定义 & Contains |
| internal/geo/loader.go | 目录扫描 + 子级加载 |
| internal/geo/geomjson.go | JSON 多边形解析 |
| internal/geo/s2util.go | S2 辅助函数 |
| internal/geo/jetgeo.go | Engine/NewEngine/Reverse |
| internal/geo/cache.go | LoadingCache 实现 |
| internal/httpx/server.go | HTTP Server + graceful |
| internal/httpx/handler_reverse.go | reverse 接口 |
| internal/httpx/middleware.go | 日志/恢复/metrics |
| internal/util/logger.go | 日志初始化 |
| internal/util/envflag.go | env + flag 组合 |
| pkg/jetgeo/api.go | 对外导出 Engine 接口 |
| docs/api.md | API 文档复制/更新 |
| docs/design.md | 设计文档 |
| testdata/polygons/... | 小样本数据 |

---
## 13. 关键实现伪代码
### 13.1 Engine.Reverse
```go
func (e *Engine) Reverse(lat,lng float64) (*GeoInfo,bool) {
  var province *RegionCache
  for _, rc := range e.province { if rc.Contains(lat,lng) { province = rc; break } }
  if province == nil { return nil,false }
  gi := &GeoInfo{Province: province.Name, ProvinceCode: province.Code, Adcode: province.Code, Level: LevelProvince}
  if e.cfg.Level >= LevelCity { // 需要 city
     cities,_ := e.cityCache.Get(province.Code)
     for _, c := range cities { if c.Contains(lat,lng) { gi.City=c.Name; gi.CityCode=c.Code; gi.Adcode=c.Code; gi.Level=LevelCity; break } }
  }
  if gi.Adcode != "" && e.cfg.Level >= LevelDistrict {
     districts,_ := e.districtCache.Get(gi.Adcode)
     for _, d := range districts { if d.Contains(lat,lng) { gi.District=d.Name; gi.DistrictCode=d.Code; gi.Adcode=d.Code; gi.Level=LevelDistrict; break } }
  }
  gi.FormatAddress = gi.Province + gi.City + gi.District
  return gi,true
}
```

### 13.2 JSON 解析
```go
type jsonPoint struct { Lat float64 `json:"lat"`; Lng float64 `json:"lng"` }
func ParsePolygons(r io.Reader) ([][]jsonPoint, error) { ... }
```

### 13.3 缓存 loader
```go
cityCache := NewLoadingCache(cfg.CityTTL, func(parent string)([]*RegionCache,error){ return LoadChildren(LevelCity,parent) })
```

---
## 14. 性能与内存考虑
- 省级区域数少, 直接线性 contains O(P)
- 市/区惰性加载减少初始内存
- 后续可加: (1) 省级 bounding box 预过滤; (2) S2 cell 前缀 -> 候选集合索引
- 可提供 Benchmark 结果 (实现后补充)

---
## 15. 后续扩展点
| 方向 | 描述 |
|------|------|
| 多数据源 | 支持增量更新/热替换数据目录 |
| 更多级别 | 添加街道/社区级目录结构 |
| 并行加载 | 初始化时并发加载省级 | 
| 索引优化 | 建立 geohash / S2 cell 前缀索引 | 
| OpenAPI | 自动生成 swagger 文档 |

---
## 16. 迁移风险与规避
| 风险 | 说明 | 缓解措施 |
|------|------|-----------|
| S2 多边形方向差异 | Loop 方向错误导致 contains 失败 | 校验 & 反转 loop |
| 大文件解析耗时 | JSON 阻塞启动 | 支持并发 + 进度日志 |
| 缓存失效策略不一致 | 与 Java 期望行为差异 | 文档声明 + 可配置 TTL |
| 浮点比较边界 | 边界点是否包含 | 允许轻微误差, 必要时扩大或使用 ContainsPoint 容差 |

---
## 17. License / 第三方依赖 (预估)
| 依赖 | 用途 | 许可证 |
|------|------|--------|
| github.com/golang/geo/s2 | S2 几何 | Apache-2.0 |
| github.com/go-chi/chi/v5 | HTTP 路由 | MIT |
| go.uber.org/zap (可选) | 日志 | MIT |
| github.com/prometheus/client_golang (可选) | Metrics | Apache-2.0 |

---
## 18. 下一步
1. 确认本设计 (你反馈)  
2. 初始化 go.mod + 目录 scaffold  
3. 实现核心解析与 Engine  
4. 加入 HTTP 接口 + 基础测试  
5. 增量优化

---
如需修改: 可指出希望采用的 Web 框架/缓存库/配置模式差异, 我会调整再继续生成脚手架代码。
