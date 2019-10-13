# rtc-api

为了更方便的运行单元测试，rtc-api能帮助你运行微服务以及它的依赖项，它将用真实的微服务api代替mock

## Getting Started(simple)

### Download

- [rtc releases](https://gitlab.p2shop.cn:8443/qa/rtc-api/-/tags)

### Start
mac
```
$ ./rtc-darwin_amd64 run mysql
```
linux
```
$ ./rtc-linux_amd64 run mysql,kafka
```
windows
```
$ ./rtc-win64.exe run mysql,kafka,sqlserver,redis
```

## Getting Started(advanced)

### Preparation

1. 微服务之间的依赖关系：来自mysql服务（127.0.0.1:3308）以及[rtc数据库](../example/settings.sql)
2. 微服务运行时的文件：来自微服务的[docker镜像](https://ci.p2shop.com.cn/)
3. 微服务运行时的数据库：来自mysql服务（127.0.0.1:3306）以及微服务运行时需要的[数据库](../example/basedata.sql)
4. 设置环境变量，启动本地api：$env:IS_RTC_API="Y" go run .
5. 恢复环境变量：$env:IS_RTC_API="N" 

### Start

mac
```
$ ./rtc-darwin_amd64 run go-api --env "staging"
```
linux
```
$ ./rtc-linux_amd64 run go-api --env "staging"
```
windows
```
$ ./rtc-win64.exe run go-api --env "staging"
```

### Service find
You can find all the servcies and support fuzzy queries.
```
$ ./rtc-api ls
go-api
inventories-api
inventories-csl-adapter
```
```
$ ./rtc-api ls in
inventories-api
inventories-csl-adapter
```


## log
https://job-monitor.p2shop.com.cn/#/

service: rtc-api

## Help

```bash
$ .\rtc-api.exe run --help
$ .\rtc-api.exe --help
$ .\rtc-api.exe ls --help
```

## References

- mysql tool: [mysql](https://github.com/go-sql-driver/mysql)
- mysql dump: [dump](https://github.com/relax-space/go-mysqldump)
- joblog: [joblog](https://github.com/ElandGroup/joblog)
- kafka tool: [kafka-go](github.com/segmentio/kafka-go)
- host tool : [goodhosts](github.com/lextoumbourou/goodhosts)
- configuration tool: 
  - [viper](https://github.com/spf13/viper) 
  - [yaml](github.com/ghodss/yaml)
  - [kingpin](github.com/alecthomas/kingpin)
- wait-for: [wait-for](https://github.com/fmiguelez/wait-for.git)
- utils: https://github.com/pangpanglabs/goutils