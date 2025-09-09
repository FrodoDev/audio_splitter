package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/xuri/excelize/v2"
)

func AddLyrics(xlsxPath string) {
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		log.Fatalf("open xlsx fail %s", xlsxPath)
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			log.Println("XlsxReader close file", err)
		}
	}()

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("lyrics")
	if err != nil {
		log.Println(err)
	}

	for i := range rows {
		if rows[i][0] != "1" {
			continue
		}

		pth := rows[i][1]
		pth = strings.ReplaceAll(pth, "homeDir", homeDir)

		lyrics := rows[i][3]

		// 打开 MP3 文件，准备写入标签
		for _, suffix := range []string{"slow", "fast"} {
			name := rows[i][2]
			ext := filepath.Ext(name)
			firstName := strings.TrimSuffix(name, ext)
			pathName := pth + suffix + "/" + firstName + "-" + suffix + ext
			// log.Println("flag;", name, ext, firstName, pathName)
			tag, err := id3v2.Open(pathName, id3v2.Options{Parse: true})
			if err != nil {
				log.Fatalf("打开文件失败: %s, %v", pathName, err)
			}
			defer tag.Close()

			// 准备纯文本歌词，语言代码为 eng（英语）
			uslt := id3v2.UnsynchronisedLyricsFrame{
				Encoding:          id3v2.EncodingUTF8,
				Language:          "eng",
				ContentDescriptor: "",
				Lyrics:            lyrics,
			}

			// 添加 USLT 帧
			tag.AddUnsynchronisedLyricsFrame(uslt)

			// 保存修改
			if err = tag.Save(); err != nil {
				log.Fatalf("保存标签失败:%s %v", pathName, err)
			}

			log.Println("歌词已成功写入 mp3 文件", pathName)
		}
	}
}
