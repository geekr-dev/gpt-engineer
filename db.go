package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type DB struct {
	path string
}

func NewDB(path string) *DB {
	absPath, _ := filepath.Abs(path)
	_ = os.MkdirAll(absPath, os.ModePerm)
	return &DB{path: absPath}
}

func (db *DB) Get(key string) (string, error) {
	content, err := ioutil.ReadFile(filepath.Join(db.path, key))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (db *DB) Set(key, val string) error {
	return ioutil.WriteFile(filepath.Join(db.path, key), []byte(val), 0644)
}

type DBs struct {
	memory    *DB
	logs      *DB
	identity  *DB
	input     *DB
	workspace *DB
}

func NewDBs(rootPath string) *DBs {
	inputPath := filepath.Join(rootPath, "example")
	memoryPath := filepath.Join(inputPath, "memory")
	workspacePath := filepath.Join(inputPath, "workspace")
	identityPath := filepath.Join(rootPath, "identity")
	return &DBs{
		memory:    NewDB(memoryPath),
		logs:      NewDB(filepath.Join(memoryPath, "logs")),
		identity:  NewDB(identityPath),
		input:     NewDB(inputPath),
		workspace: NewDB(workspacePath),
	}
}
