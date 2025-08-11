<div align="center">

# ![resources/logo/lionchat.png](resources/logo/logo.png)
(⌛ 正在开发中.)
English | [简体中文]
chat lion 是一个采用 Go 技术栈构建即时通讯后端系统，使用 Gin、GORM、Redis、WebSocket 和 Kafka，实现了一个功能丰富的聊天应用。
[documentation]() | 
[前端项目](https://github.com/kapbl/LionChat-Fronted)
[后端项目](https://github.com/kapbl/LionChat)
[测试服务器运行指标的项目](https://github.com/kapbl/Lion-Chat-Test)
</div>


## 🎯 特点
- 支持消息分片✅
- 分层架构✅
- 工作池模式✅
- 好友管理✅
- 单聊和群聊✅
- 文字消息/语音消息/文件消息✅
- 支持分布式部署❌
- 语音聊天✅
- 视频聊天✅
- AI聊天❌
- 此刻(类似朋友圈)✅
- 聊天记录备份❌
- 实时语音转录+情感分析❌
- 跨语言无障碍沟通❌
- 对话摘要与决策提炼❌
- Docker部署✅
## 🎐本地开发
- Go 1.24+
- gin
- GORM
- nginx
- docker
## 🎐Docker Compose 部署
- 构建镜像
```bash
docker-compose build
```
- 运行容器
```bash
docker-compose up -d
```
## 🦁画廊
### 服务端架构
![服务端架构](resources/logo/Untitled-2025-08-07-1051.png)
### 客户端之间的通信过程
![客户端之间的通信过程](resources/logo/客户端之间的通信过程.svg)
### 1. 两个好友在聊天
好友A:
![聊天1](resources/assest/57d8e366a96b0678301d3c98df8eea4a.png)
好友B:
![聊天2](resources/assest/7ee1812a213af185fca6a3a361148511.png)
### 2. 两个好友在语音电话
![聊天1](resources/assest/4c20b36be80f9d92ed6b98bfdb1558ab.png)
![聊天2](resources/assest/2ca35e8a20cefe905b77c1ba4407d9fb.png)
![聊天2](resources/assest/de143ed179263b8084b09d438c5db8ce.png)
