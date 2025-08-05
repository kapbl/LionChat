package main

import (
	"cchat/internal/service"
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"fmt"
	"time"
)

func main() {
	// 初始化日志
	logger.InitLogger()

	fmt.Println("=== ChatLion 消息分片功能演示 ===")

	// 创建分片管理器
	fm := service.NewFragmentManager()
	defer fm.Stop()

	// 演示1: 小消息（不需要分片）
	fmt.Println("\n1. 测试小消息（不需要分片）")
	smallMsg := &protocol.Message{
		Avatar:       "user1.jpg",
		FromUsername: "Alice",
		From:         "user-001",
		To:           "user-002",
		Content:      "Hello, Bob! How are you today?",
		ContentType:  1, // 文本消息
		Type:         "text",
		MessageType:  1, // 单聊
	}

	if fm.ShouldFragment(smallMsg) {
		fmt.Println("❌ 小消息被错误地标记为需要分片")
	} else {
		fmt.Println("✅ 小消息正确地不需要分片")
	}

	// 演示2: 大消息（需要分片）
	fmt.Println("\n2. 测试大消息（需要分片）")

	// 创建一个大文件内容（超过64KB）
	largeContent := make([]byte, service.MaxFragmentSize+10000)
	for i := range largeContent {
		largeContent[i] = byte('A' + (i % 26))
	}

	largeMsg := &protocol.Message{
		Avatar:       "user1.jpg",
		FromUsername: "Alice",
		From:         "user-001",
		To:           "user-002",
		Content:      "Sending you a large file",
		ContentType:  2, // 文件消息
		Type:         "file",
		MessageType:  1, // 单聊
		Url:          "large_file.txt",
		FileSuffix:   "txt",
		File:         largeContent,
	}

	if !fm.ShouldFragment(largeMsg) {
		fmt.Println("❌ 大消息应该被分片但没有被标记")
		return
	}

	fmt.Printf("✅ 大消息需要分片（大小: %d 字节）\n", len(largeContent))

	// 演示3: 消息分片
	fmt.Println("\n3. 执行消息分片")
	fragments, err := fm.FragmentMessage(largeMsg)
	if err != nil {
		fmt.Printf("❌ 分片失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 消息成功分片为 %d 个片段\n", len(fragments))
	fmt.Printf("   消息ID: %s\n", fragments[0].MessageId)
	fmt.Printf("   校验和: %s\n", fragments[0].Checksum)

	// 显示每个分片的信息
	for i, fragment := range fragments {
		fmt.Printf("   分片 %d: 大小 %d 字节\n", i, len(fragment.File))
	}

	// 演示4: 模拟网络传输和消息重组
	fmt.Println("\n4. 模拟消息重组")

	// 创建接收端分片管理器
	receiveFM := service.NewFragmentManager()
	defer receiveFM.Stop()

	// 模拟乱序接收分片
	order := []int{1, 0, 2} // 假设有3个分片，乱序接收
	if len(fragments) >= 3 {
		fmt.Println("   模拟乱序接收分片...")

		var reassembledMsg *protocol.Message
		for i, idx := range order {
			if idx >= len(fragments) {
				continue
			}

			fmt.Printf("   接收分片 %d/%d\n", idx, len(fragments)-1)

			msg, isComplete, err := receiveFM.ProcessFragment(fragments[idx])
			if err != nil {
				fmt.Printf("❌ 处理分片失败: %v\n", err)
				return
			}

			if isComplete {
				fmt.Printf("✅ 消息重组完成！（处理了 %d 个分片）\n", i+1)
				reassembledMsg = msg
				break
			} else {
				fmt.Printf("   等待更多分片...\n")
			}
		}

		// 验证重组结果
		if reassembledMsg != nil {
			fmt.Println("\n5. 验证重组结果")
			if reassembledMsg.Content == largeMsg.Content {
				fmt.Println("✅ 消息内容匹配")
			} else {
				fmt.Println("❌ 消息内容不匹配")
			}

			if len(reassembledMsg.File) == len(largeMsg.File) {
				fmt.Printf("✅ 文件大小匹配（%d 字节）\n", len(reassembledMsg.File))
			} else {
				fmt.Printf("❌ 文件大小不匹配（原始: %d, 重组: %d）\n",
					len(largeMsg.File), len(reassembledMsg.File))
			}

			// 验证文件内容
			contentMatch := true
			for i := 0; i < len(largeMsg.File) && i < len(reassembledMsg.File); i++ {
				if largeMsg.File[i] != reassembledMsg.File[i] {
					contentMatch = false
					break
				}
			}

			if contentMatch {
				fmt.Println("✅ 文件内容完全匹配")
			} else {
				fmt.Println("❌ 文件内容不匹配")
			}

			if !reassembledMsg.IsFragmented {
				fmt.Println("✅ 重组后的消息正确地标记为非分片")
			} else {
				fmt.Println("❌ 重组后的消息仍然标记为分片")
			}
		}
	} else {
		// 如果分片数少于3个，按顺序处理
		fmt.Println("   按顺序接收分片...")

		var reassembledMsg *protocol.Message
		for i, fragment := range fragments {
			fmt.Printf("   接收分片 %d/%d\n", i, len(fragments)-1)

			msg, isComplete, err := receiveFM.ProcessFragment(fragment)
			if err != nil {
				fmt.Printf("❌ 处理分片失败: %v\n", err)
				return
			}

			if isComplete {
				fmt.Printf("✅ 消息重组完成！\n")
				reassembledMsg = msg
				break
			}
		}

		if reassembledMsg != nil {
			fmt.Println("\n5. 验证重组结果")
			fmt.Println("✅ 消息重组成功")
		}
	}

	// 演示6: 性能统计
	fmt.Println("\n6. 性能统计")
	fmt.Printf("   最大分片大小: %d KB\n", service.MaxFragmentSize/1024)
	fmt.Printf("   分片超时时间: %v\n", service.FragmentTimeout)
	fmt.Printf("   最大分片数量: %d\n", service.MaxFragments)
	fmt.Printf("   当前待处理分片组: %d\n", fm.GetPendingFragmentsCount())

	// 演示7: 清理测试
	fmt.Println("\n7. 测试分片清理")
	testMsg := &protocol.Message{
		MessageId:      "test-cleanup",
		IsFragmented:   true,
		FragmentIndex:  0,
		TotalFragments: 2,
		Timestamp:      time.Now().Unix(),
		Checksum:       "test-checksum",
		From:           "user-001",
		To:             "user-002",
	}

	_, _, err = fm.ProcessFragment(testMsg)
	if err != nil {
		fmt.Printf("❌ 处理测试分片失败: %v\n", err)
	} else {
		fmt.Printf("✅ 创建了测试分片组，待处理分片组数: %d\n", fm.GetPendingFragmentsCount())
	}

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("\n消息分片功能特性:")
	fmt.Println("• 自动检测大消息并进行分片")
	fmt.Println("• 支持乱序分片接收和重组")
	fmt.Println("• 消息完整性校验（MD5）")
	fmt.Println("• 自动清理过期分片")
	fmt.Println("• 支持各种消息类型（文本、文件、图片等）")
	fmt.Println("• 线程安全的并发处理")
}
