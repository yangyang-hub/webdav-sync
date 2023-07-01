# webdav-sync

#### 介绍

定时从webdav同步数据到本地，主要用于视频文件等大文件


#### 使用说明

项目初始化 go.mod, 在当前工程目录下依次执行以下两个命令

1. go mod init
2. go mod tidy
3. 修改docker-compose.yml中webdav地址，以及本地磁盘地址映射，同步时间使用corn表达式配置
4. docker-compose up -d
