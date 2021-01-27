# http-server
## 启动参数启动

|       参数      | 默认值 |             说明            |
|-----------------|--------|-----------------------------|
| -port           | 18080  | 监听端口                    |
| -debug          | false  | 是否为debug模式             |
| -static.dir     |        | 静态文件目录                |
| -static.history | false  | 静态文件是否使用history路由 |
| -proxy.uri      | /proxy | 接口代理根路径              |
| -proxy.target   |        | 接口代理目标路径            |
| -proxy.forward  | false  | 接口代理是否做forward转发   |

> proxy.uri和proxy.target有任意一个为空时不启用接口代理

## 配置文件启动
> 配置文件会默认读取当前路径下的conf.json,  
> 如果不存在会读取http-server执行文件所在目录下的conf.json  
> 使用配置文件时支持代理多个接口  
> 使用配置文件时支持访问频率限制  
> crtFile和keyFile都不为空时使用https监听  

|        参数       |               说明               |
|-------------------|----------------------------------|
| port              | 监听端口                         |
| crtFile           | 证书crt公钥                      |
| keyFile           | 证书key文件                      |
| debug             | 是否为debug模式                  |
| static.dir        | 静态文件目录                     |
| static.history    | 静态文件是否使用history路由      |
| limit.enable      | 是否启用访问限制                 |
| limit.delay       | 访问接口是延迟多少毫秒           |
| limit.period      | IP访问限制周期                   |
| limit.times       | IP访问限制一个周期内可访问的次数 |
| limit.message     | 超过访问次数返回的消息           |
| limit.whiteIps[]  | IP访问限制白名单                 |
| proxies[].uri     | 接口代理根路径                   |
| proxies[].target  | 接口代理目标路径                 |
| proxies[].forward | 接口代理是否做forward转发        |

配置文件示例
```json
{
  "port": 18081,
  "crtFile": "",
  "keyFile": "",
  "debug": true,
  "static": {
    "dir": "./public",
    "history": true
  },
  "limit": {
    "enable": true,
    "delay": 0,
    "period": 2,
    "times": 10,
    "message": "{\"error\": \"access_limit\",\"error_message\": \"Access limit!\"}",
    "whiteIps": [
      "192.168.0.101"
    ]
  },
  "proxies": [
    {
      "uri": "/rpc",
      "target": "http://192.168.0.102:8080",
      "forward": true
    }
  ]
}
```
