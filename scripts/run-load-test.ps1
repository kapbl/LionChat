# LionChat 负载测试启动脚本
# 用于Windows PowerShell环境

param(
    [int]$Clients = 1000,
    [int]$Duration = 300,
    [string]$ServerURL = "ws://localhost/ws",
    [int]$Concurrency = 100,
    [int]$MessageRate = 100,
    [string]$JWTKey = "chat-loadtest-key-2024",
    [switch]$Simple,
    [switch]$Monitor,
    [switch]$Help
)

# 显示帮助信息
if ($Help) {
    Write-Host "LionChat 负载测试脚本" -ForegroundColor Green
    Write-Host ""
    Write-Host "参数说明:" -ForegroundColor Yellow
    Write-Host "  -Clients      客户端数量 (默认: 1000)"
    Write-Host "  -Duration     测试持续时间(秒) (默认: 300)"
    Write-Host "  -ServerURL    服务器地址 (默认: ws://localhost/ws)"
    Write-Host "  -Concurrency  并发连接数 (默认: 100)"
    Write-Host "  -MessageRate  每秒消息数 (默认: 100)"
    Write-Host "  -JWTKey       JWT签名密钥"
    Write-Host "  -Simple       使用简单测试模式"
    Write-Host "  -Monitor      启动监控模式"
    Write-Host "  -Help         显示此帮助信息"
    Write-Host ""
    Write-Host "使用示例:" -ForegroundColor Yellow
    Write-Host "  .\run-load-test.ps1 -Clients 10000 -Duration 600"
    Write-Host "  .\run-load-test.ps1 -Simple -Clients 100"
    Write-Host "  .\run-load-test.ps1 -Monitor"
    exit 0
}

# 检查Go环境
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "错误: 未找到Go环境，请先安装Go" -ForegroundColor Red
    exit 1
}

# 设置工作目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$TestDir = Join-Path $ProjectRoot "test"

Set-Location $ProjectRoot

Write-Host "=== LionChat 负载测试 ===" -ForegroundColor Green
Write-Host "项目目录: $ProjectRoot"
Write-Host "测试目录: $TestDir"
Write-Host ""

# 监控模式
if ($Monitor) {
    Write-Host "启动监控模式..." -ForegroundColor Yellow
    
    # 启动系统监控
    Start-Job -ScriptBlock {
        while ($true) {
            $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
            $cpu = Get-Counter "\Processor(_Total)\% Processor Time" | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue
            $memory = Get-Counter "\Memory\Available MBytes" | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue
            
            Write-Host "[$timestamp] CPU: $([math]::Round($cpu, 2))% | 可用内存: $([math]::Round($memory, 0))MB" -ForegroundColor Cyan
            Start-Sleep 5
        }
    } -Name "SystemMonitor"
    
    # 启动服务器监控
    Start-Job -ScriptBlock {
        param($ServerURL)
        $baseURL = $ServerURL -replace "ws://", "http://" -replace "wss://", "https://" -replace "/ws", ""
        
        while ($true) {
            try {
                $response = Invoke-RestMethod -Uri "$baseURL/api/monitor/server" -Method Get -TimeoutSec 5
                if ($response.code -eq 0) {
                    $data = $response.data
                    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
                    Write-Host "[$timestamp] 连接数: $($data.client_count) | Goroutines: $($data.goroutine_count) | 内存: $([math]::Round($data.memory_stats.Alloc/1024/1024, 2))MB" -ForegroundColor Green
                }
            } catch {
                Write-Host "无法连接到服务器监控API" -ForegroundColor Red
            }
            Start-Sleep 10
        }
    } -ArgumentList $ServerURL -Name "ServerMonitor"
    
    Write-Host "监控已启动，按任意键停止..." -ForegroundColor Yellow
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    
    # 停止监控任务
    Get-Job -Name "SystemMonitor", "ServerMonitor" | Stop-Job
    Get-Job -Name "SystemMonitor", "ServerMonitor" | Remove-Job
    
    Write-Host "监控已停止" -ForegroundColor Yellow
    exit 0
}

# 检查测试目录
if (-not (Test-Path $TestDir)) {
    Write-Host "创建测试目录..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $TestDir -Force | Out-Null
}

Set-Location $TestDir

# 编译测试程序
if ($Simple) {
    Write-Host "编译简单测试程序..." -ForegroundColor Yellow
    $TestFile = "simple_load_test.go"
    $TestExe = "simple_load_test.exe"
} else {
    Write-Host "编译完整测试程序..." -ForegroundColor Yellow
    $TestFile = "load_test.go"
    $TestExe = "load_test.exe"
}

if (-not (Test-Path $TestFile)) {
    Write-Host "错误: 测试文件 $TestFile 不存在" -ForegroundColor Red
    exit 1
}

# 编译
try {
    Write-Host "正在编译 $TestFile..." -ForegroundColor Yellow
    go build -o $TestExe $TestFile
    if ($LASTEXITCODE -ne 0) {
        throw "编译失败"
    }
    Write-Host "编译成功" -ForegroundColor Green
} catch {
    Write-Host "编译失败: $_" -ForegroundColor Red
    exit 1
}

# 显示测试配置
Write-Host ""
Write-Host "=== 测试配置 ===" -ForegroundColor Yellow
Write-Host "服务器地址: $ServerURL"
Write-Host "客户端数量: $Clients"
Write-Host "测试时长: $Duration 秒"
if (-not $Simple) {
    Write-Host "并发连接: $Concurrency"
    Write-Host "消息速率: $MessageRate/秒"
    Write-Host "JWT密钥: $JWTKey"
}
Write-Host "测试模式: $(if ($Simple) { '简单模式' } else { '完整模式' })"
Write-Host ""

# 确认开始测试
Write-Host "按任意键开始测试，或按 Ctrl+C 取消..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

Write-Host ""
Write-Host "=== 开始负载测试 ===" -ForegroundColor Green
$StartTime = Get-Date

# 运行测试
try {
    if ($Simple) {
        # 简单测试模式
        & ".\$TestExe"
    } else {
        # 完整测试模式
        & ".\$TestExe" -url $ServerURL -clients $Clients -rate $MessageRate -duration "$($Duration)s" -concurrency $Concurrency -jwt-key $JWTKey
    }
    
    $EndTime = Get-Date
    $Duration = $EndTime - $StartTime
    
    Write-Host ""
    Write-Host "=== 测试完成 ===" -ForegroundColor Green
    Write-Host "开始时间: $($StartTime.ToString('yyyy-MM-dd HH:mm:ss'))"
    Write-Host "结束时间: $($EndTime.ToString('yyyy-MM-dd HH:mm:ss'))"
    Write-Host "总耗时: $([math]::Round($Duration.TotalSeconds, 2)) 秒"
    
} catch {
    Write-Host "测试执行失败: $_" -ForegroundColor Red
    exit 1
}

# 清理
Write-Host ""
Write-Host "清理测试文件..." -ForegroundColor Yellow
if (Test-Path $TestExe) {
    Remove-Item $TestExe -Force
}

Write-Host "测试完成！" -ForegroundColor Green

# 提示查看日志
Write-Host ""
Write-Host "提示:" -ForegroundColor Yellow
Write-Host "- 查看服务器日志: Get-Content ..\logs\loadtest.log -Tail 50"
Write-Host "- 查看服务器状态: Invoke-RestMethod -Uri 'http://localhost/api/monitor/server'"
Write-Host "- 查看客户端列表: Invoke-RestMethod -Uri 'http://localhost/api/monitor/clients'"