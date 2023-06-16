package main

import (
	"regexp"
	"strings"
)

type fileItem struct {
	path string
	code string
}

func parseChat(chat string) []fileItem {
	regex := regexp.MustCompile("```(.*?)```")
	matches := regex.FindAllStringSubmatch(chat, -1)

	files := make([]fileItem, 0, len(matches))

	for _, match := range matches {
		path := strings.Split(match[1], "\n")[0]
		code := strings.Join(strings.Split(match[1], "\n")[1:], "\n")
		files = append(files, fileItem{path, code})
	}

	return files
}

func toFiles(chat string, workspace *DB) {
	workspace.Set("all_output.txt", chat)

	files := parseChat(chat)
	for _, file := range files {
		workspace.Set(file.path, file.code)
	}
}
