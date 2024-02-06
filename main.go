package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"tobiasthedanish/typeagen/config"
	"tobiasthedanish/typeagen/json"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	cfg, err := config.ParseConfig()
	check(err)
	// fmt.Println(cfg.ToString())

	var (
		output io.Writer
	)
	if cfg.Output == "" {
		output = os.Stdout
	} else {
		output, err = os.OpenFile(cfg.Output, os.O_RDWR|os.O_CREATE, fs.ModePerm)
		check(err)
	}

	for _, epc := range cfg.Endpoints {

		res, err := config.Fetch(epc.Url, config.RequestOptions{Method: "GET"})
		check(err)

		jsonStr, err := config.GetBodyAsString(res)
		check(err)

		obj, arr := json.ParseJson(jsonStr)
		if obj != nil {
			// fmt.Println(obj.ToString(0))
			genInterfaces(output, *obj, epc.RootType)
		} else if arr != nil {
			// fmt.Println(arr.ToString(0))

			if arr.Properties[0].IsObject() {
				genInterfaces(output, arr.Properties[0].(json.JsonObject), epc.RootType)
			}
		} else {
			fmt.Println("Something serious went wrong when parsing json!")
		}
	}
}

func genInterfaces(out io.Writer, obj json.JsonObject, root string) {
	types := []json.JsonObject{}
	obj.Key = root
	types = genInterface(out, obj, types)

	for len(types) > 0 {
		popped := types[len(types)-1]
		types = types[:len(types)-1]
		types = genInterface(out, popped, types)
	}
}

func genInterface(out io.Writer, obj json.JsonObject, types []json.JsonObject) []json.JsonObject {
	fmt.Fprintln(out, "interface "+obj.GetType()+" {")
	for key := range obj.Properties {
		if obj.Properties[key].IsObject() {
			types = append(types, obj.Properties[key].(json.JsonObject))
		}

		fmt.Fprint(out, "  "+obj.Properties[key].ToTypeString())
	}
	fmt.Fprintln(out, "}")

	return types
}
