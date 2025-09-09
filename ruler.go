package main

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ruleReader interface {
	Read() ([]Chunk, error)
	// Valid() (bool, error)
	Decode()
}

type segment struct {
	startTime  string
	endTime    string
	suffixName string // new audio file suffix name
	outputPath string
}

// fix correct h:m:s to 00:00:00
// if endTime is 0, correct to audio file life
func (s *segment) fix() {
	// s.startTime
}

type Chunk struct {
	InputPath string
	Name      string // source audio file name
	SegList   []segment
}

// XlsxReader read xlsx in fix style:
// 1. sheet name must be esl
// 2. title's format is <input path, file name, output path, [start time, end time, suffix name]...>
// 3. time's format is h:m:s, 00:02:56 means 2 minutes and 56 seconds
// 4. output path can omit, defauts value is the same with input path
// 5. suffix name can omit, defauts value is "file name" base name + "-start time" + extension,
// eg: a.mp3 default new file name is a-00:02:56.mp3
// with suffix name "slow", get a-slow.mp3
type XlsxReader struct {
	path  string
	rows  [][]string
	chunk []Chunk
}

var _ ruleReader = (*XlsxReader)(nil)

func (x *XlsxReader) Read() ([]Chunk, error) {
	f, err := excelize.OpenFile(x.path)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println("XlsxReader close file", err)
		}
	}()

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("esl")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	x.rows = rows[1:]

	x.Decode()
	// x.Print()

	return x.chunk, nil
}

// Valid return isValid, invalid reason
func (x *XlsxReader) Valid(i int) (valid bool, reason error) {
	row := x.rows[i]
	if len(row) != 4 {
		reason = fmt.Errorf("row %d not equal 5", i)
		return
	}

	// todo support other audio file "wav"
	if !strings.Contains(row[1], "mp3") {
		reason = fmt.Errorf("row %d offer a audio file which is not supported", i)
		return
	}

	r2 := strings.Split(row[2], ",")
	if len(r2) != 4 {
		reason = fmt.Errorf("row %d column 2 is not match the format [start time,endtime,suffix name,output path]", i)
		return
	}

	r3 := strings.Split(row[3], ",")
	if len(r3) != 4 {
		reason = fmt.Errorf("row %d column 3 is not match the format [start time,endtime,suffix name,output path]", i)
		return
	}

	valid = true
	return
}

func (x *XlsxReader) Decode() {
	x.chunk = make([]Chunk, 0, len(x.rows))
	for i := range x.rows {
		valid, err := x.Valid(i)
		if !valid {
			fmt.Println(err)
			continue
		}

		chunk := Chunk{}
		chunk.InputPath = x.rows[i][0]
		chunk.InputPath = strings.ReplaceAll(chunk.InputPath, "homeDir", homeDir)
		chunk.Name = x.rows[i][1]
		s1s := strings.Split(x.rows[i][2], ",")
		op := strings.ReplaceAll(s1s[3], "homeDir", homeDir)
		chunk.SegList = append(chunk.SegList,
			segment{startTime: s1s[0], endTime: s1s[1], suffixName: s1s[2], outputPath: op})

		s2s := strings.Split(x.rows[i][3], ",")
		op = strings.ReplaceAll(s2s[3], "homeDir", homeDir)
		chunk.SegList = append(chunk.SegList,
			segment{startTime: s2s[0], endTime: s2s[1], suffixName: s2s[2], outputPath: op})

		x.chunk = append(x.chunk, chunk)
	}
}

func (x *XlsxReader) Print() {
	for i := range x.chunk {
		fmt.Println(i, x.chunk[i].InputPath, x.chunk[i].Name, x.chunk[i].SegList)
	}
}

func ReadRuler() {

}
