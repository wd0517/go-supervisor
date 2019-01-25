# go-supervisor
[Practice project] Implemented supervisor with Go

## 简单使用

```shell
$ go build -o supervisord go-supervisor
$ ./supervisord -echo_conf > config.yml
$ ./supervisord -c config.yml    # start
$ ./supervisord -s shutdown      # stop(graceful)
$ ./supervisord -s stop          # stop
```

## 配置文件说明:

#### Supervisor配置:

```yml
supervisord:
  logfile: "supervisor.log"     # 日志名称
  pidfile: ".supervisor.pid"    # daemon状态pid文件名称
  logpath: "logs"               # 日志路径
  httpserver: "127.0.0.1:8080"  # http服务监听地址及端口
  nodaemon: false               # 开启方式, 默认为守护进程
```

#### 普通进程配置:
```yml
- name: fileServer
  command: python -m SimpleHTTPServer 8003
  directory: /your/directory
  autostart: false       # 是否默认启动
  autorerestart: false   # 进程崩溃后是否自动重启
  startsecs: 0           # 进程启动多少秒后认为已经成功运行
  startretries: 0        # 进程启动失败后的尝试次数
```

### 对外提供的接口

1. 所有进程状态总览 http://127.0.0.1:8080/
2. 根据名称启动某个进程 http://127.0.0.1:8080/start/?name=processname
3. 根据名称停止某个进程 http://127.0.0.1:8080/stop/?name=processname
4. 根据名称重启某个进程 http://127.0.0.1:8080/restart/?name=processname

## TODOS
- [ ] 日志切割
- [ ] 支持Reload配置 [./supervisor -s reload]
- [ ] 支持进程组
- [ ] 支持对系统资源的限制(cgroup)
