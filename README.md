# FaileEsl
关于freeswitch 
针对sip攻击做的防火墙处理微服务
linux环境
需要fail2ban
安装并配置

windows环境
用的自带的防火墙

通过监听freeswitch的事件
如果发现注册失败，就把本次失败的ip添加到一个临时集合中
如果本次ip尝试的次数大于5次，就禁止他的访问
