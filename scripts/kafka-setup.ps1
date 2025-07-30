# Kafka环境设置脚本
# 用于Windows PowerShell环境

Write-Host "=== ChatLion Kafka环境设置 ===" -ForegroundColor Green

# 检查Docker是否安装
function Test-Docker {
    try {
        docker --version | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# 检查Docker Compose是否安装
function Test-DockerCompose {
    try {
        docker-compose --version | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# 启动Kafka环境
function Start-KafkaEnvironment {
    Write-Host "启动Kafka环境..." -ForegroundColor Yellow
    
    # 停止可能存在的容器
    docker-compose -f docker-compose.kafka.yml down
    
    # 启动新的环境
    docker-compose -f docker-compose.kafka.yml up -d
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Kafka环境启动成功!" -ForegroundColor Green
        Write-Host "等待服务完全启动..." -ForegroundColor Yellow
        Start-Sleep -Seconds 30
        
        # 检查服务状态
        docker-compose -f docker-compose.kafka.yml ps
        
        Write-Host ""
        Write-Host "=== 服务访问地址 ===" -ForegroundColor Cyan
        Write-Host "Kafka Broker: localhost:9092"
        Write-Host "Kafka UI: http://localhost:8080"
        Write-Host "Kafka Manager: http://localhost:9000"
        Write-Host "Zookeeper: localhost:2181"
        Write-Host ""
    }
    else {
        Write-Host "Kafka环境启动失败!" -ForegroundColor Red
        exit 1
    }
}

# 停止Kafka环境
function Stop-KafkaEnvironment {
    Write-Host "停止Kafka环境..." -ForegroundColor Yellow
    docker-compose -f docker-compose.kafka.yml down
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Kafka环境已停止!" -ForegroundColor Green
    }
    else {
        Write-Host "停止Kafka环境失败!" -ForegroundColor Red
    }
}

# 查看Kafka日志
function Show-KafkaLogs {
    Write-Host "显示Kafka日志..." -ForegroundColor Yellow
    docker-compose -f docker-compose.kafka.yml logs -f kafka
}

# 测试Kafka连接
function Test-KafkaConnection {
    Write-Host "测试Kafka连接..." -ForegroundColor Yellow
    
    # 列出主题
    docker exec chatLion-kafka kafka-topics --list --bootstrap-server localhost:9092
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Kafka连接测试成功!" -ForegroundColor Green
    }
    else {
        Write-Host "Kafka连接测试失败!" -ForegroundColor Red
    }
}

# 运行应用测试
function Test-Application {
    Write-Host "运行应用Kafka测试..." -ForegroundColor Yellow
    
    # 确保在项目根目录
    if (!(Test-Path "go.mod")) {
        Write-Host "请在项目根目录运行此脚本!" -ForegroundColor Red
        exit 1
    }
    
    # 运行测试
    go test -v ./test/kafka/...
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "应用测试通过!" -ForegroundColor Green
    }
    else {
        Write-Host "应用测试失败!" -ForegroundColor Red
    }
}

# 清理Kafka数据
function Clear-KafkaData {
    Write-Host "清理Kafka数据..." -ForegroundColor Yellow
    
    # 停止服务
    docker-compose -f docker-compose.kafka.yml down
    
    # 删除数据卷
    docker volume rm chatLion_kafka-data chatLion_zookeeper-data chatLion_zookeeper-logs 2>$null
    
    Write-Host "Kafka数据已清理!" -ForegroundColor Green
}

# 显示帮助信息
function Show-Help {
    Write-Host ""
    Write-Host "=== ChatLion Kafka设置脚本 ===" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "用法: .\scripts\kafka-setup.ps1 [命令]" -ForegroundColor White
    Write-Host ""
    Write-Host "可用命令:" -ForegroundColor Yellow
    Write-Host "  start     - 启动Kafka环境"
    Write-Host "  stop      - 停止Kafka环境"
    Write-Host "  restart   - 重启Kafka环境"
    Write-Host "  logs      - 查看Kafka日志"
    Write-Host "  test      - 测试Kafka连接"
    Write-Host "  app-test  - 运行应用测试"
    Write-Host "  clean     - 清理Kafka数据"
    Write-Host "  status    - 查看服务状态"
    Write-Host "  help      - 显示此帮助信息"
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Cyan
    Write-Host "  .\scripts\kafka-setup.ps1 start"
    Write-Host "  .\scripts\kafka-setup.ps1 test"
    Write-Host ""
}

# 检查服务状态
function Show-Status {
    Write-Host "=== Kafka服务状态 ===" -ForegroundColor Cyan
    docker-compose -f docker-compose.kafka.yml ps
    
    Write-Host ""
    Write-Host "=== Docker容器状态 ===" -ForegroundColor Cyan
    docker ps --filter "name=chatLion-"
}

# 主逻辑
if (!(Test-Docker)) {
    Write-Host "错误: 未找到Docker，请先安装Docker Desktop" -ForegroundColor Red
    exit 1
}

if (!(Test-DockerCompose)) {
    Write-Host "错误: 未找到Docker Compose，请确保Docker Desktop已正确安装" -ForegroundColor Red
    exit 1
}

# 解析命令行参数
switch ($args[0]) {
    "start" {
        Start-KafkaEnvironment
    }
    "stop" {
        Stop-KafkaEnvironment
    }
    "restart" {
        Stop-KafkaEnvironment
        Start-Sleep -Seconds 5
        Start-KafkaEnvironment
    }
    "logs" {
        Show-KafkaLogs
    }
    "test" {
        Test-KafkaConnection
    }
    "app-test" {
        Test-Application
    }
    "clean" {
        Clear-KafkaData
    }
    "status" {
        Show-Status
    }
    "help" {
        Show-Help
    }
    default {
        if ($args.Count -eq 0) {
            Show-Help
        }
        else {
            Write-Host "未知命令: $($args[0])" -ForegroundColor Red
            Show-Help
        }
    }
}