# rtc-api

To make unit testing easier, rtc-api can help you run microservices and its dependencies, it will replace the mock with a real microservice api.

## 1. Getting Started

### Download rtc

- [github](https://github.com/relax-space/rtc-api/releases)
- [gitlab](https://gitlab.p2shop.cn:8443/qa/rtc-api/-/tags)

### Rename

- rename `rtc-darwin_amd64` to `rtc`

### Start(simple)
```
$ ./rtc run mysql,kafka,sqlserver,redis
```

### Start(advanced)

download example file from [github](https://github.com/relax-space/rtc-api/releases) or [gitlab](https://gitlab.p2shop.cn:8443/qa/rtc-api/-/tags)

```
$ docker-compose -f example/docker-compose.yml up -d
$ ./rtc run go-api --env local --image-env="" --docker-no-log
```

### Service find
You can find all the servcies and support fuzzy queries.
```
$ ./rtc ls
go-api
inventories-api
inventories-csl-adapter
```
```
$ ./rtc ls in
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
- kafka tool: [kafka-go](https://github.com/segmentio/kafka-go)
- host tool : [goodhosts](https://github.com/lextoumbourou/goodhosts)
- configuration tool: 
  - [viper](https://github.com/spf13/viper) 
  - [yaml](https://github.com/ghodss/yaml)
  - [kingpin](https://github.com/alecthomas/kingpin)
- wait-for: [wait-for](https://github.com/fmiguelez/wait-for.git)
- utils: https://github.com/pangpanglabs/goutils

## 4. Dependencies

- [docker](https://docker_practice.gitee.io/us_en/install/)
- [docker-compose](https://docker_practice.gitee.io/us_en/compose/install.html)