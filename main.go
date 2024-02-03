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
		"https://jsonplaceholder.typicode.com/users/1",
		RequestOptions{Method: "GET"},
	)

	check(err)
	defer res.Body.Close()

	body := ""
	scanner := bufio.NewScanner(res.Body)
	for i := 0; scanner.Scan() && i < 5; i++ {
		body = body + scanner.Text()
	}

	return body
}

func main() {
	json := getJson()
	i := 0
	fmt.Println(json)

	for i < len(json) {
		switch json[i] {
		case '{':
			parseObject(json, i, "")
		}
	}
}

type JsonObject struct {
	Key        string
	Properties map[string]JsonObject
}

func parseObject(json string, i int, key string) (JsonObject, int) {
	i = skipWhiteSpace(json, i+1)
	res := JsonObject{Key: key, Properties: make(map[string]JsonObject)}

	for i < len(json) {
		propName, i := readKey(json, i)
		fmt.Println(propName)
		i = skipWhiteSpace(json, i)
		if json[i] != ':' {
			panic("Malformed json.")
		}
		i += 1

		i = skipWhiteSpace(json, i)

		if unicode.IsNumber(rune(json[i])) {
			panic("TODO: parseNumber not implemented")
		} else if unicode.IsLetter(rune(json[i])) {
			panic("TODO: parseBoolean not implemented")
		}

		switch json[i] {
		case '"':
			panic("TODO: parseString not implemented")
			break

		case '{':
			obj := JsonObject{}
			obj, i = parseObject(json, i, propName)
			res.Properties[key] = obj
			break

		case '[':
			panic("TODO: parseArray not implemented")
			break
		}

		panic("Unreachable in parseObject switch")
	}

	return res, i
}

func readKey(json string, i int) (string, int) {
	i = skipWhiteSpace(json, i)
	if json[i] != '"' {
		panic(fmt.Sprintf("Expected start of key at position %d", i))
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

	for unicode.IsSpace(rune(json[i])) {
		i += 1
	}

	return i
}
