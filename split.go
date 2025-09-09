package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func SplitAudio(xlsxPath string) {
	reader := &XlsxReader{path: xlsxPath}
	chunks, err := reader.Read()
	if err != nil {
		fmt.Println("read fail")
		return
	}
	_ = chunks

	for i := range chunks {
		for j := range chunks[i].SegList {
			if err := SplitAudioBySeg(chunks[i].InputPath, chunks[i].Name, chunks[i].SegList[j]); err != nil {
				fmt.Printf("切割失败: %v\n", err)
				return
			}
			// fmt.Println("切割成功:", chunks[i].InputPath, chunks[i].Name, chunks[i].SegList[j])
		}

	}
}

func SplitAudioBySeg(srcPath, srcName string, seg segment) error {
	inputFile := path.Join(srcPath, srcName)
	// 检查输入文件是否存在
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("输入文件不存在%s", inputFile)
	}

	// 创建输出目录
	if _, err := os.Stat(seg.outputPath); os.IsNotExist(err) {
		if err1 := os.MkdirAll(seg.outputPath, 0755); err1 != nil {
			return fmt.Errorf("创建输出目录失败: %v", err1)
		}
	}

	startTime, err := correctTime(inputFile, seg.startTime, false)
	if err != nil {
		return err
	}

	endTime, err := correctTime(inputFile, seg.endTime, true)
	if err != nil {
		return err
	}

	ext := filepath.Ext(srcName)
	baseName := strings.TrimSuffix(filepath.Base(srcName), ext)
	destName := fmt.Sprintf("%s-%s%s", baseName, seg.suffixName, ext)
	outputPath := path.Join(seg.outputPath, destName)

	// 构建ffmpeg命令
	cmd := exec.Command(
		"ffmpeg",
		"-i", inputFile,
		"-ss", startTime,
		"-to", endTime,
		"-c", "copy", // 不重新编码，处理速度快且不损失质量
		outputPath,
		"-y", // 覆盖已存在的文件
	)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("切割失败: %v, 输出: %s", err, string(output))
	}

	fmt.Printf("已生成片段: %s\n", outputPath)
	return nil
}

func correctTime(srcFile, srcTime string, isEnd bool) (string, error) {
	if isEnd && srcTime == "0" {
		endTime, err := getDurationString(srcFile)
		if err != nil {
			return "", err
		}

		return subtractSecondsFromTime(endTime, 39)
	}

	return subtractSecondsFromTime(srcTime, 3)
}

// 时间字符串减去n秒，返回新的时间字符串
func subtractSecondsFromTime(timeStr string, n int) (string, error) {
	// 解析时间字符串为总秒数
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("时间格式错误，应为HH:MM:SS")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("解析小时错误: %w", err)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("解析分钟错误: %w", err)
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("解析秒错误: %w", err)
	}

	// 转换为总秒数
	totalSeconds := hours*3600 + minutes*60 + seconds

	// 减去n秒（确保不小于0）
	newTotalSeconds := totalSeconds - n
	if newTotalSeconds < 0 {
		newTotalSeconds = 0 // 处理负数情况，默认为0
	}

	// 转换回HH:MM:SS格式
	newHours := newTotalSeconds / 3600
	newMinutes := (newTotalSeconds % 3600) / 60
	newSeconds := newTotalSeconds % 60

	// 格式化输出，确保两位数
	return fmt.Sprintf("%02d:%02d:%02d", newHours, newMinutes, newSeconds), nil
}

// 获取格式为 "HH:MM:SS" 的时长字符串
func getDurationString(srcFile string) (string, error) {
	// 调用ffprobe获取带毫秒的时分秒格式（如 00:21:34.79）
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-sexagesimal", // 输出 HH:MM:SS.sss 格式
		"-of", "default=noprint_wrappers=1:nokey=1",
		srcFile,
	)

	// 执行命令并获取结果
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取时长失败：%w", err)
	}

	// 处理结果：去掉空格和毫秒部分
	durationWithMs := strings.TrimSpace(string(output)) // 得到 "00:21:34.79"
	parts := strings.Split(durationWithMs, ".")         // 按 "." 分割成 ["00:21:34", "79"]
	if len(parts) < 1 {
		return "", fmt.Errorf("时长格式错误：%s", durationWithMs)
	}

	return parts[0], nil // 返回 "00:21:34"
}

// SplitAudioByDuration 按指定时长切割音频
// inputPath: 输入音频路径
// outputDir: 输出目录
// durationSec: 每个片段的时长（秒）
func SplitAudioByDuration(inputPath, outputDir string, durationSec int) error {
	// 检查输入文件是否存在
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return errors.New("输入文件不存在")
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 获取输入文件信息
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)

	// 获取音频总时长（秒）
	totalDuration, err := getAudioDuration(inputPath)
	if err != nil {
		return fmt.Errorf("获取音频时长失败: %v", err)
	}

	// 计算需要切割的片段数量
	numChunks := (totalDuration + durationSec - 1) / durationSec

	// 执行切割
	for i := 0; i < numChunks; i++ {
		startTime := i * durationSec
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_chunk_%d%s", baseName, i+1, ext))

		// 构建ffmpeg命令
		// -i: 输入文件
		// -ss: 开始时间（秒）
		// -t: 持续时间（秒）
		// -c: 编码方式（复制，不重新编码）
		cmd := exec.Command(
			"ffmpeg",
			"-i", inputPath,
			"-ss", strconv.Itoa(startTime),
			"-t", strconv.Itoa(durationSec),
			"-c", "copy",
			outputPath,
			"-y", // 覆盖已存在的文件
		)

		// 执行命令
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("切割失败 (片段 %d): %v, 输出: %s", i+1, err, string(output))
		}

		fmt.Printf("已生成片段: %s\n", outputPath)
	}

	return nil
}

// SplitAudioByTime 按指定时间段切割音频
// inputPath: 输入音频路径
// outputPath: 输出音频路径
// startTime: 开始时间（格式: 秒 或 时:分:秒）
// endTime: 结束时间（格式同上）
func SplitAudioByTime(inputPath, outputPath, startTime, endTime string) error {
	// 检查输入文件是否存在
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return errors.New("输入文件不存在")
	}

	// 创建输出目录
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 构建ffmpeg命令
	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-ss", startTime,
		"-to", endTime,
		"-c", "copy", // 不重新编码，处理速度快且不损失质量
		outputPath,
		"-y", // 覆盖已存在的文件
	)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("切割失败: %v, 输出: %s", err, string(output))
	}

	fmt.Printf("已生成片段: %s\n", outputPath)
	return nil
}

// 获取音频总时长（秒）
func getAudioDuration(filePath string) (int, error) {
	// 使用ffprobe获取音频信息
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// 解析时长（秒）
	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return int(duration), nil
}
