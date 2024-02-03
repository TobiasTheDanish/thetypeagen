package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"unicode"
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

func getJson() string {
	res, err := fetch(
		"https://jsonplaceholder.typicode.com/users",
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
	json := getJson()
	i := 0
	// fmt.Println(json)

	for i < len(json) {
		switch json[i] {
		case '{':
			res := JsonObject{}
			res, i = parseObject(json, i, "")
			fmt.Println(res)
		case '[':
			res := JsonArray{}
			res, i = parseArray(json, i, "")
			fmt.Println(res)
		}

		i = skipWhiteSpace(json, i)
	}
}

type JsonElement interface {
	GetKey() string
	GetValue() (string, map[string]JsonElement, []JsonElement)
}

type JsonPrimitive struct {
	Key   string
	Value string
}

func (p JsonPrimitive) GetKey() string {
	return p.Key
}

func (p JsonPrimitive) GetValue() (string, map[string]JsonElement, []JsonElement) {
	return p.Value, nil, nil
}

type JsonObject struct {
	Key        string
	Properties map[string]JsonElement
}

func (o JsonObject) GetKey() string {
	return o.Key
}

func (o JsonObject) GetValue() (string, map[string]JsonElement, []JsonElement) {
	return "", o.Properties, nil
}

type JsonArray struct {
	Key        string
	Properties []JsonElement
}

func (a JsonArray) GetKey() string {
	return a.Key
}

func (a JsonArray) GetValue() (string, map[string]JsonElement, []JsonElement) {
	return "", nil, a.Properties
}

func parsePrimitive(json string, i int, key string) (JsonPrimitive, int) {
	res := JsonPrimitive{Key: key}
	value := []byte{}
	if unicode.IsNumber(rune(json[i])) {
		for ; i < len(json) && unicode.IsNumber(rune(json[i])); i++ {
			value = append(value, json[i])
		}
		res.Value = string(value)
		return res, i

	} else if unicode.IsLetter(rune(json[i])) {
		panic("TODO: parse boolean or null not implemented")
	} else if json[i] == '"' {
		i += 1 // skip first "

		for ; i < len(json) && json[i] != '"'; i++ {
			value = append(value, json[i])
		}
		res.Value = string(value)

		i += 1 // skip last "
		return res, i

	} else {
		panic("Malformed json in parsePrimitive.")
	}
}

func parseObject(json string, i int, key string) (JsonObject, int) {
	i = skipWhiteSpace(json, i+1)
	res := JsonObject{Key: key, Properties: make(map[string]JsonElement)}

	for i < len(json) && json[i] != '}' {
		propName := ""
		propName, i = readKey(json, i)
		i = skipWhiteSpace(json, i)
		if json[i] != ':' {
			panic("Malformed json in parseObject.")
		}
		i += 1

		i = skipWhiteSpace(json, i)

		switch json[i] {
		case '{':
			obj := JsonObject{}
			obj, i = parseObject(json, i, propName)
			res.Properties[propName] = obj
			break

		case '[':
			obj := JsonArray{}
			obj, i = parseArray(json, i, propName)
			res.Properties[propName] = obj
			break

		default:
			p := JsonPrimitive{}
			p, i = parsePrimitive(json, i, propName)
			res.Properties[propName] = p
		}

		if json[i] == ',' {
			i += 1
		}
		// panic("")
	}

	if json[i] == '}' {
		i += 1
	}

	return res, i
}

func parseArray(json string, i int, key string) (JsonArray, int) {
	i = skipWhiteSpace(json, i+1)
	res := JsonArray{Key: key, Properties: []JsonElement{}}

	for i < len(json) && json[i] != ']' {
		i = skipWhiteSpace(json, i)

		if json[i] == '{' {
			obj := JsonObject{}
			obj, i = parseObject(json, i, "")
			res.Properties = append(res.Properties, obj)
		} else {
			p := JsonPrimitive{}
			p, i = parsePrimitive(json, i, "")
			res.Properties = append(res.Properties, p)
		}

		if json[i] == ',' {
			i += 1
		}
		// panic("")
	}

	return res, i
}

func readKey(json string, i int) (string, int) {
	i = skipWhiteSpace(json, i)
	if json[i] != '"' {
		panic(fmt.Sprintf("Expected start of key at position %d, found %c", i, json[i]))
	}
	i += 1

	key := []byte{}
	for json[i] != '"' {
		key = append(key, json[i])
		i += 1
	}

	return string(key), i + 1
}

func skipWhiteSpace(json string, i int) int {

	for i < len(json) && unicode.IsSpace(rune(json[i])) {
		i += 1
	}

	return i
}
