# LionChat 系统优化脚本
# 用于优化Windows系统以支持1万个并发连接

param(
    [switch]$Apply,
    [switch]$Check,
    [switch]$Restore,
    [switch]$Help
)

# 需要管理员权限
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "此脚本需要管理员权限，请以管理员身份运行PowerShell" -ForegroundColor Red
    exit 1
}

# 显示帮助信息
if ($Help) {
    Write-Host "LionChat 系统优化脚本" -ForegroundColor Green
    Write-Host ""
    Write-Host "参数说明:" -ForegroundColor Yellow
    Write-Host "  -Apply    应用系统优化设置"
    Write-Host "  -Check    检查当前系统配置"
    Write-Host "  -Restore  恢复默认系统设置"
    Write-Host "  -Help     显示此帮助信息"
    Write-Host ""
    Write-Host "使用示例:" -ForegroundColor Yellow
    Write-Host "  .\optimize-system.ps1 -Check"
    Write-Host "  .\optimize-system.ps1 -Apply"
    Write-Host "  .\optimize-system.ps1 -Restore"
    exit 0
}

# 备份注册表键值
function Backup-RegistryKey {
    param(
        [string]$KeyPath,
        [string]$ValueName,
        [string]$BackupFile
    )
    
    try {
        $value = Get-ItemProperty -Path $KeyPath -Name $ValueName -ErrorAction SilentlyContinue
        if ($value) {
            "$KeyPath|$ValueName|$($value.$ValueName)" | Out-File -FilePath $BackupFile -Append
        } else {
            "$KeyPath|$ValueName|NOT_EXIST" | Out-File -FilePath $BackupFile -Append
        }
    } catch {
        "$KeyPath|$ValueName|ERROR" | Out-File -FilePath $BackupFile -Append
    }
}

# 恢复注册表键值
function Restore-RegistryKey {
    param(
        [string]$BackupFile
    )
    
    if (-not (Test-Path $BackupFile)) {
        Write-Host "备份文件不存在: $BackupFile" -ForegroundColor Red
        return
    }
    
    Get-Content $BackupFile | ForEach-Object {
        $parts = $_ -split '\|'
        if ($parts.Length -eq 3) {
            $keyPath = $parts[0]
            $valueName = $parts[1]
            $originalValue = $parts[2]
            
            try {
                if ($originalValue -eq "NOT_EXIST") {
                    Remove-ItemProperty -Path $keyPath -Name $valueName -ErrorAction SilentlyContinue
                    Write-Host "已删除: $keyPath\$valueName" -ForegroundColor Yellow
                } elseif ($originalValue -ne "ERROR") {
                    Set-ItemProperty -Path $keyPath -Name $valueName -Value $originalValue
                    Write-Host "已恢复: $keyPath\$valueName = $originalValue" -ForegroundColor Green
                }
            } catch {
                Write-Host "恢复失败: $keyPath\$valueName" -ForegroundColor Red
            }
        }
    }
}

# 检查当前系统配置
function Check-SystemConfiguration {
    Write-Host "=== 当前系统配置检查 ===" -ForegroundColor Green
    Write-Host ""
    
    # 检查TCP连接限制
    Write-Host "TCP连接配置:" -ForegroundColor Yellow
    try {
        $tcpParams = Get-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" -ErrorAction SilentlyContinue
        
        $maxUserPort = if ($tcpParams.MaxUserPort) { $tcpParams.MaxUserPort } else { "默认(5000)" }
        $tcpTimedWaitDelay = if ($tcpParams.TcpTimedWaitDelay) { $tcpParams.TcpTimedWaitDelay } else { "默认(240)" }
        $maxHashTableSize = if ($tcpParams.MaxHashTableSize) { $tcpParams.MaxHashTableSize } else { "默认(512)" }
        
        Write-Host "  MaxUserPort: $maxUserPort"
        Write-Host "  TcpTimedWaitDelay: $tcpTimedWaitDelay"
        Write-Host "  MaxHashTableSize: $maxHashTableSize"
    } catch {
        Write-Host "  无法读取TCP参数" -ForegroundColor Red
    }
    
    # 检查内存配置
    Write-Host ""
    Write-Host "内存配置:" -ForegroundColor Yellow
    $memory = Get-CimInstance -ClassName Win32_ComputerSystem
    $totalMemory = [math]::Round($memory.TotalPhysicalMemory / 1GB, 2)
    Write-Host "  总内存: $totalMemory GB"
    
    $availableMemory = Get-Counter "\Memory\Available MBytes" | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue
    Write-Host "  可用内存: $([math]::Round($availableMemory / 1024, 2)) GB"
    
    # 检查网络配置
    Write-Host ""
    Write-Host "网络配置:" -ForegroundColor Yellow
    $networkAdapters = Get-NetAdapter | Where-Object { $_.Status -eq "Up" }
    foreach ($adapter in $networkAdapters) {
        Write-Host "  网卡: $($adapter.Name) - $($adapter.LinkSpeed)"
    }
    
    # 检查防火墙状态
    Write-Host ""
    Write-Host "防火墙状态:" -ForegroundColor Yellow
    try {
        $firewallProfiles = Get-NetFirewallProfile
        foreach ($profile in $firewallProfiles) {
            Write-Host "  $($profile.Name): $($profile.Enabled)"
        }
    } catch {
        Write-Host "  无法获取防火墙状态" -ForegroundColor Red
    }
    
    # 检查进程限制
    Write-Host ""
    Write-Host "进程信息:" -ForegroundColor Yellow
    $processes = Get-Process
    Write-Host "  当前进程数: $($processes.Count)"
    
    $handles = ($processes | Measure-Object -Property HandleCount -Sum).Sum
    Write-Host "  总句柄数: $handles"
    
    Write-Host ""
    Write-Host "=== 检查完成 ===" -ForegroundColor Green
}

# 应用系统优化
function Apply-SystemOptimization {
    Write-Host "=== 应用系统优化设置 ===" -ForegroundColor Green
    Write-Host ""
    
    $backupFile = "$env:TEMP\lionchat_system_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss').txt"
    Write-Host "创建备份文件: $backupFile" -ForegroundColor Yellow
    
    # 备份当前设置
    $tcpParamsPath = "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters"
    
    Write-Host "备份当前设置..." -ForegroundColor Yellow
    Backup-RegistryKey -KeyPath $tcpParamsPath -ValueName "MaxUserPort" -BackupFile $backupFile
    Backup-RegistryKey -KeyPath $tcpParamsPath -ValueName "TcpTimedWaitDelay" -BackupFile $backupFile
    Backup-RegistryKey -KeyPath $tcpParamsPath -ValueName "MaxHashTableSize" -BackupFile $backupFile
    Backup-RegistryKey -KeyPath $tcpParamsPath -ValueName "TcpNumConnections" -BackupFile $backupFile
    
    Write-Host "优化TCP连接参数..." -ForegroundColor Yellow
    
    try {
        # 增加最大用户端口数
        Set-ItemProperty -Path $tcpParamsPath -Name "MaxUserPort" -Value 65534 -Type DWord
        Write-Host "  设置 MaxUserPort = 65534" -ForegroundColor Green
        
        # 减少TIME_WAIT状态的等待时间
        Set-ItemProperty -Path $tcpParamsPath -Name "TcpTimedWaitDelay" -Value 30 -Type DWord
        Write-Host "  设置 TcpTimedWaitDelay = 30" -ForegroundColor Green
        
        # 增加TCP连接哈希表大小
        Set-ItemProperty -Path $tcpParamsPath -Name "MaxHashTableSize" -Value 65536 -Type DWord
        Write-Host "  设置 MaxHashTableSize = 65536" -ForegroundColor Green
        
        # 增加最大TCP连接数
        Set-ItemProperty -Path $tcpParamsPath -Name "TcpNumConnections" -Value 16777214 -Type DWord
        Write-Host "  设置 TcpNumConnections = 16777214" -ForegroundColor Green
        
    } catch {
        Write-Host "  TCP参数设置失败: $_" -ForegroundColor Red
    }
    
    # 优化网络缓冲区
    Write-Host ""
    Write-Host "优化网络缓冲区..." -ForegroundColor Yellow
    
    try {
        # 设置接收窗口自动调整
        netsh int tcp set global autotuninglevel=normal
        Write-Host "  启用TCP接收窗口自动调整" -ForegroundColor Green
        
        # 启用TCP Chimney Offload
        netsh int tcp set global chimney=enabled
        Write-Host "  启用TCP Chimney Offload" -ForegroundColor Green
        
        # 启用接收端缩放
        netsh int tcp set global rss=enabled
        Write-Host "  启用接收端缩放(RSS)" -ForegroundColor Green
        
    } catch {
        Write-Host "  网络优化失败: $_" -ForegroundColor Red
    }
    
    # 优化内存管理
    Write-Host ""
    Write-Host "优化内存管理..." -ForegroundColor Yellow
    
    try {
        $memoryPath = "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management"
        
        # 备份内存管理设置
        Backup-RegistryKey -KeyPath $memoryPath -ValueName "LargeSystemCache" -BackupFile $backupFile
        Backup-RegistryKey -KeyPath $memoryPath -ValueName "SystemPages" -BackupFile $backupFile
        
        # 优化系统缓存
        Set-ItemProperty -Path $memoryPath -Name "LargeSystemCache" -Value 1 -Type DWord
        Write-Host "  启用大系统缓存" -ForegroundColor Green
        
        # 增加系统页表项
        Set-ItemProperty -Path $memoryPath -Name "SystemPages" -Value 0 -Type DWord
        Write-Host "  优化系统页表" -ForegroundColor Green
        
    } catch {
        Write-Host "  内存优化失败: $_" -ForegroundColor Red
    }
    
    # 优化文件系统
    Write-Host ""
    Write-Host "优化文件系统..." -ForegroundColor Yellow
    
    try {
        $filesystemPath = "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem"
        
        # 备份文件系统设置
        Backup-RegistryKey -KeyPath $filesystemPath -ValueName "NtfsDisableLastAccessUpdate" -BackupFile $backupFile
        
        # 禁用NTFS最后访问时间更新
        Set-ItemProperty -Path $filesystemPath -Name "NtfsDisableLastAccessUpdate" -Value 1 -Type DWord
        Write-Host "  禁用NTFS最后访问时间更新" -ForegroundColor Green
        
    } catch {
        Write-Host "  文件系统优化失败: $_" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "=== 优化完成 ===" -ForegroundColor Green
    Write-Host "备份文件已保存到: $backupFile" -ForegroundColor Yellow
    Write-Host "请重启系统以使所有更改生效" -ForegroundColor Red
    Write-Host ""
    Write-Host "重启后建议运行以下命令验证设置:" -ForegroundColor Yellow
    Write-Host "  .\optimize-system.ps1 -Check"
}

# 恢复系统设置
function Restore-SystemSettings {
    Write-Host "=== 恢复系统设置 ===" -ForegroundColor Green
    Write-Host ""
    
    # 查找备份文件
    $backupFiles = Get-ChildItem -Path $env:TEMP -Filter "lionchat_system_backup_*.txt" | Sort-Object LastWriteTime -Descending
    
    if ($backupFiles.Count -eq 0) {
        Write-Host "未找到备份文件" -ForegroundColor Red
        return
    }
    
    Write-Host "找到以下备份文件:" -ForegroundColor Yellow
    for ($i = 0; $i -lt $backupFiles.Count; $i++) {
        Write-Host "  [$i] $($backupFiles[$i].Name) - $($backupFiles[$i].LastWriteTime)"
    }
    
    $selection = Read-Host "请选择要恢复的备份文件编号 (0-$($backupFiles.Count-1))"
    
    try {
        $selectedFile = $backupFiles[[int]$selection]
        Write-Host "恢复备份文件: $($selectedFile.FullName)" -ForegroundColor Yellow
        
        Restore-RegistryKey -BackupFile $selectedFile.FullName
        
        # 恢复网络设置
        Write-Host ""
        Write-Host "恢复网络设置..." -ForegroundColor Yellow
        netsh int tcp set global autotuninglevel=normal
        netsh int tcp set global chimney=disabled
        netsh int tcp set global rss=enabled
        
        Write-Host ""
        Write-Host "=== 恢复完成 ===" -ForegroundColor Green
        Write-Host "请重启系统以使所有更改生效" -ForegroundColor Red
        
    } catch {
        Write-Host "恢复失败: $_" -ForegroundColor Red
    }
}

# 主逻辑
Write-Host "LionChat 系统优化工具" -ForegroundColor Green
Write-Host "当前用户: $env:USERNAME"
Write-Host "系统版本: $((Get-CimInstance Win32_OperatingSystem).Caption)"
Write-Host ""

if ($Check) {
    Check-SystemConfiguration
} elseif ($Apply) {
    Write-Host "警告: 此操作将修改系统注册表设置" -ForegroundColor Red
    Write-Host "建议在测试环境中先进行验证" -ForegroundColor Yellow
    Write-Host ""
    $confirm = Read-Host "确认要应用系统优化吗? (y/N)"
    
    if ($confirm -eq 'y' -or $confirm -eq 'Y') {
        Apply-SystemOptimization
    } else {
        Write-Host "操作已取消" -ForegroundColor Yellow
    }
} elseif ($Restore) {
    Write-Host "警告: 此操作将恢复之前的系统设置" -ForegroundColor Red
    Write-Host ""
    $confirm = Read-Host "确认要恢复系统设置吗? (y/N)"
    
    if ($confirm -eq 'y' -or $confirm -eq 'Y') {
        Restore-SystemSettings
    } else {
        Write-Host "操作已取消" -ForegroundColor Yellow
    }
} else {
    Write-Host "请指定操作参数，使用 -Help 查看帮助" -ForegroundColor Yellow
}