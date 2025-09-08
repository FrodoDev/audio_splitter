package main

import (
	"bytes"
	"fmt"

	"io/ioutil"

	"github.com/mikkyang/id3-go"
	uni "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// todo 不能获取到 macOS 上 音乐 app 和 Musicolet 上歌词的部分
// github.com/mikkyang/id3-go not work
// github.com/dhowden/tag not work

// 读取并合并所有歌词帧
func ReadLyrics(path string) (string, error) {
	file, err := id3.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var fullLyrics string

	for _, frame := range file.AllFrames() {
		id := frame.Id()
		if id == "ULT" || id == "USLT" || id == "COMM" || id == "SYLT" {
			data := frame.Bytes()
			content, err := parseLyricsFrame(data)
			if err != nil {
				fmt.Printf("帧 %s 解析失败: %v\n", id, err)
			} else {
				fmt.Printf("帧 %s 内容:\n%s\n\n", id, content)
			}
		}
	}

	return fullLyrics, nil
}

// 解析单个歌词帧内容：跳过编码字节、语言码、描述字符串，解码歌词正文
func parseLyricsFrame(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("frame data too short")
	}
	encoding := data[0]
	language := data[1:4]
	fmt.Println("language:", string(language))
	rest := data[4:]

	descEnd := 0
	for i := 0; i < len(rest)-1; i += 2 {
		if rest[i] == 0x00 && rest[i+1] == 0x00 {
			descEnd = i + 2
			break
		}
	}
	lyricsData := rest[descEnd:]

	var decoder transform.Transformer

	switch encoding {
	case 0:
		return string(lyricsData), nil
	case 1:
		decoder = uni.UTF16(uni.LittleEndian, uni.ExpectBOM).NewDecoder()
	default:
		return "", fmt.Errorf("unsupported encoding: %d", encoding)
	}

	reader := bytes.NewReader(lyricsData)
	transformedReader := transform.NewReader(reader, decoder)
	decodedBytes, err := ioutil.ReadAll(transformedReader)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

// func main() {
//     lyrics, err := ReadLyrics("ESL0001-Introducing-Yourself.mp3")
//     if err != nil {
//         fmt.Println("Error:", err)
//         return
//     }
//     fmt.Println("完整歌词:\n", lyrics)
// }
