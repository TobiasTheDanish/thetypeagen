package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"tobiasthedanish/typeagen/json"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type RequestOptions struct {
	Method  string
	Headers http.Header
	Body    io.Reader
}

func fetch(url string, options RequestOptions) (*http.Response, error) {
	req, err := http.NewRequest(options.Method, url, options.Body)
	if err != nil {
		return nil, err
	}

	req.Header = options.Headers
	return http.DefaultClient.Do(req)
}

func getJson(url string) string {
	res, err := fetch(
		url,
		RequestOptions{Method: "GET"},
	)

	check(err)
	defer res.Body.Close()

	body := ""
	scanner := bufio.NewScanner(res.Body)
	for i := 0; scanner.Scan(); i++ {
		body = body + scanner.Text()
	}

	return body
}

func main() {
	url := os.Args[1]
	root := os.Args[2]
	jsonStr := getJson(url)
	// fmt.Println(json)

	obj, arr := json.ParseJson(jsonStr)
	if obj != nil {
		fmt.Println(obj.ToString(0))
		genInterfaces(*obj, root)
	} else if arr != nil {
		fmt.Println(arr.ToString(0))

		if arr.Properties[0].IsObject() {
			genInterfaces(arr.Properties[0].(json.JsonObject), root)
		}
	}

}

func genInterfaces(obj json.JsonObject, root string) {
	types := []json.JsonObject{}
	obj.Key = root
	types = genInterface(obj, types)

	for len(types) > 0 {
		popped := types[len(types)-1]
		types = types[:len(types)-1]
		types = genInterface(popped, types)
	}
}

func genInterface(obj json.JsonObject, types []json.JsonObject) []json.JsonObject {
	fmt.Println("interface " + obj.GetType() + " {")
	for key := range obj.Properties {
		if obj.Properties[key].IsObject() {
			types = append(types, obj.Properties[key].(json.JsonObject))
		}

		fmt.Print("  " + obj.Properties[key].ToTypeString())
	}
	fmt.Println("}")

	return types
}
