package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	startList = []string{"time", "year", "people", "way", "day", "thing"}
	wordList  = []string{"life", "world", "school", "state", "family", "student", "group", "country", "problem", "hand", "part", "place", "case", "week", "company", "system", "program", "work", "government", "number", "night", "point", "home", "water", "room", "mother", "area", "money", "story", "fact", "month", "lot", "right", "study", "book", "eye", "job", "word", "business", "issue", "side", "kind", "head", "house", "service", "friend", "father", "power", "hour", "game", "line", "end", "member", "law", "car", "city", "community", "name", "president", "team", "minute", "idea", "kid", "body", "information", "back", "face", "others", "level", "office", "door", "health", "person", "art", "war", "history", "party", "result", "change", "morning", "reason", "research", "moment", "air", "teacher", "force", "education"}
	extList   = []string{"txt", "md", "pdf", "jpg", "jpeg", "png", "mp4", "mp3", "csv"}
	startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate   = time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	dateList  = []time.Time{}
	wordIdx   = 0
	extIdx    = 0
	dateIdx   = 0
)

func nextWord() string {
	word := wordList[wordIdx%len(wordList)]
	wordIdx++
	return word
}

func nextExt() string {
	ext := extList[extIdx%len(extList)]
	extIdx++
	return ext
}

func setDate(filename string, r int) {
	date := dateList[dateIdx%len(dateList)]
	m := 17 * dateIdx / len(dateList)
	date = date.Add(time.Duration(m) * time.Hour)
	dateIdx++
	os.Chtimes(filename, date, date)
}

func genFile(dir string, a int) {
	os.MkdirAll(dir, 0755)
	for i := 1; i <= 5; i++ {
		size := a*i*wordIdx*100 + extIdx
		file := nextWord() + "-" + nextWord()

		if i%3 == 0 {
			file += "-" + nextWord()
		}

		file += "." + nextExt()
		path := filepath.Join(dir, file)
		ioutil.WriteFile(path, make([]byte, size), 0644)
		setDate(path, size*size)
	}
}

func genDir(root string) {
	for _, start := range startList {

		for i := 1; i <= 5; i++ {
			dir := filepath.Join(root, start, nextWord())
			genFile(dir, 1)

			if wordIdx%3 == 0 {
				dir = filepath.Join(dir, nextWord())
				genFile(dir, 1)
			}
		}
	}
}

func main() {
	root := "/tmp/zfind"

	var c int64 = 50
	interval := (int64)(endDate.Sub(startDate).Seconds()) / c
	for i := range make([]int64, c) {
		dateList = append(dateList, startDate.Add(time.Duration(interval*(int64)(i))*time.Second))
	}

	if err := os.RemoveAll(root); err == nil {
		genDir(filepath.Join(root, "root"))
		fmt.Println("Ready.")
	} else {
		fmt.Println("Failed to clean")
	}
}
