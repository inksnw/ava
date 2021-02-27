# ava
一个去中心化的分布式任务运行平台
#### 2.0升级计划(未开始)
- [ ] 换用https://github.com/nhooyr/websocket实现
- [ ] 使用单线路ws连接实现反向代理 https://github.com/genshen/wssocks




#### 2. 启动方式
##### D的运行
会监听本机4000端口用于接收web命令,连接多台运行节点

```bash
main config.josn
```

##### H的运行

H监听端口 websocket: 4560, socks5: 4562  
会读取程序运行目录的下级文档目录下的launcher1.json,同步到管理节点
```bash
./main
```
launcher1.json结构说明
```bash
{
    "worker": "gather_spider"       ----任务标识 
    "command": "python3 deal.py",   ----执行命令行
    "dir": "/home/ubuntu/deploy/gather_spider"   ---执行环境
}
```

```


##### 3. 部分api
web状态查看
```bash
http://127.0.0.1:4000
```
发送命令请求: POST  http://127.0.0.1:4000/exectask
```bash
{
"worker": "gather_spider",
"task_id": "uuidxxxxxxxxx",
"params": "eyJtZXRob2QiOiAiZmFrZS5lY2hvIiwgInBhcmFtcyI6IHsiYSI6IDEyM319"
"route": "192.168.169.128"   ---(可选,定点投送)
}

```
实际生成的运行参数
python3 deal.py placeholder /home/ubuntu/deploy/gather_spider/{params的base64写的文件}


在运行节点上,挂127.0.0.1:4562的socks5代理,可直接穿透到内网,白名单为config.json里的配置




##### 4 bug
- [ ] concurrent write to websocket connection
- [ ] ws读取超时