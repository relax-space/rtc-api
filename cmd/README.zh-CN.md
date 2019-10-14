# rtc-api

为了更方便的运行单元测试，rtc-api能帮助你运行微服务以及它的依赖项，它将用真实的微服务api代替mock

## 1. Getting Started

### Download

- [github](https://github.com/relax-space/rtc-api/releases)
- [gitlab](https://gitlab.p2shop.cn:8443/qa/rtc-api/-/tags)

### Rename

- rename `rtc-darwin_amd64` to `rtc`

### Start(simple)

```
$ ./rtc run mysql,kafka,sqlserver,redis
```

### Start(advanced)

```
$ docker-compose -f example/docker-compose.yml up -d
$ ./rtc run go-api --env staging --image-env="" --no-log
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

- [docker](https://yeasy.gitbooks.io/docker_practice/install/)
- [docker-compose](https://yeasy.gitbooks.io/docker_practice/compose/install.html)