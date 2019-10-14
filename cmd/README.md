# rtc-api

To make unit testing easier, rtc-api can help you run microservices and its dependencies, it will replace the mock with a real microservice api.

## 1. Getting Started

### Download

- [rtc releases](https://gitlab.p2shop.cn:8443/qa/rtc-api/-/tags)

### Start(simple)
```
$ ./rtc run mysql,kafka,sqlserver,redis
```

### Start(advanced)

```
$ cd rtc-api/example
$ docker-compose up -d
$ ./rtc-api run go-api --env staging --image-env="" --no-log
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

## 2. log
https://job-monitor.p2shop.com.cn/#/

service: rtc-api

## 3. References

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