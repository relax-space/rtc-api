# rtc-api

To make unit testing easier, rtc-api can help you run microservices and its dependencies, it will replace the mock with a real microservice api.

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

1. Dependencies between microservices: from the mysql service（127.0.0.1:3308）And [rtc database](../example/settings.sql)
2. Microservice Runtime Files: [docker image](https://ci.p2shop.com.cn/) from microservices
3. Microservices database: [database](../example/basedata.sql) from mysql service (127.0.0.1:3306)
4. Start local api：$env:IS_RTC_API="Y" go run .
5. Start rtc perpare：$env:IS_RTC_API="N" 

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