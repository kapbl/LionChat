
```plaintext
your_project/
├── cmd/                # 每个可执行程序一个子目录（main 入口）
│   └── admin/          # 例如：main.go表示开启后台管理服务
│       └── main.go|
|   └── app/            # 例如：main.go表示开启app服务
│       └── main.go
| 
├── internal/               # 项目私有逻辑
│   ├── service/            # 业务逻辑层（如 UserService）
│   ├── dto/            # HTTP handler 层（或 controller）
│   └── dao/                # 数据访问层（如数据库、Redis）
├── pkg/                    # 可被其他项目复用的公共库（类似工具库）
│   └── config/              # 工具包（如加密、日志封装）
├── api/                    # API 定义（OpenAPI/Swagger, Protobuf 等）
│   └── v1/                 # v1 版本 API 接口定义
├── config/                 # 配置文件（如 yaml/json/toml）
│   └── config.yaml
├── migrations/             # 数据库迁移脚本（可配合 goose, migrate 等工具）
│   └── 001_create_user.sql
├── static/                    # 用户的头像，暂时存储在这里
├── test/                   # 项目的测试代码
│   └── yourapp_test.go
├── scripts/                # 运维或自动化脚本（如构建、部署）
│   └── build.sh
├── go.mod                  # Go modules 配置文件
├── go.sum
├── .env                    # 环境变量文件（如数据库配置）
├── .gitignore
└── README.md
```

### 扩展事件类型

参考 `internal/handler/` 中的事件处理器实现，添加新的事件类型处理逻辑。
