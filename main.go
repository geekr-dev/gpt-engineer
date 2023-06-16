package main

import (
	"encoding/json"
	"flag"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

var (
	model       string
	lang        string
	temperature float64
)

func init() {
	flag.StringVar(&model, "model", defaultModel, "The model to use")
	flag.Float64Var(&temperature, "temperature", defaultTemperature, "The temperature to use")
	flag.StringVar(&lang, "lang", defaultLang, "The language to use")
}

func main() {
	flag.Parse()

	ai := NewAI(model, temperature, lang)

	rootPath, _ := filepath.Abs("./")
	dbs := NewDBs(rootPath)

	for _, step := range STEPS {
		messages := step(ai, dbs)

		pc := runtime.FuncForPC(reflect.ValueOf(step).Pointer())
		funcName := strings.ReplaceAll(pc.Name(), "main.", "") // 去除包名
		contents, _ := json.Marshal(messages)
		dbs.logs.Set(funcName, string(contents))
	}
}
