## Go 做压测的核心优势

| 特性 | 压测价值 |
|------|---------|
| **goroutine 高并发** | 单机轻松百万级并发 |
| **零依赖** | 单个 exe，直接发同事 |
| **低内存** | 比 Java 压测工具省 90% 内存 |
| **精确计时** | 纳秒级时间精度 |
| **丰富生态** | 已有成熟的压测库 |

## 功能规划

**基础功能**
```
- GET/POST/PUT/DELETE 请求
- Header/Body 自定义
- 超时控制
- 错误统计
```

**高级功能**
```
- 并发梯度测试（10→100→1000 并发）
- 请求速率控制（RPS 限制）
- SSL/TLS 支持
- Cookie/Token 自动携带
- 环境变量引用
```

**报告输出**
```
- 响应时间分布（P50/P90/P95/P99）
- TPS/QPS 统计
- 成功/失败率
- 可视化图表（HTML 报告）
- CSV 导出
```

## 技术选型参考

```
HTTP 客户端   → net/http + httpx
并发控制      → goroutine + semaphore (golang.org/x/sync/semaphore)
速率控制      → token bucket
报告生成      → 纯 Go 生成 HTML/CSV
CLI 框架      → cobra / urfave/cli
```
实现：
- 支持并发参数化
- 自动统计 P95/P99
- 输出简洁的统计报告