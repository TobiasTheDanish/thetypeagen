package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	json := getJson(url)
	i := 0
	// fmt.Println(json)

	i = skipWhiteSpace(json, i)
	switch json[i] {
	case '{':
		res := JsonObject{}
		res, i = parseObject(json, i, "")
		fmt.Println(res.ToString(0))
		genInterfaces(res, root)
		break

	case '[':
		res := JsonArray{}
		res, i = parseArray(json, i, "")
		fmt.Println(res.ToString(0))

		if res.Properties[0].IsObject() {
			genInterfaces(res.Properties[0].(JsonObject), root)
		}
		break
	}
}

func genInterfaces(obj JsonObject, root string) {
	types := []JsonObject{}
	obj.Key = root
	types = genInterface(obj, types)

	for len(types) > 0 {
		popped := types[len(types)-1]
		types = types[:len(types)-1]
		types = genInterface(popped, types)
	}
}

func genInterface(obj JsonObject, types []JsonObject) []JsonObject {
	fmt.Println("interface " + obj.GetType() + " {")
	for key := range obj.Properties {
		if obj.Properties[key].IsObject() {
			types = append(types, obj.Properties[key].(JsonObject))
		}

		fmt.Print("  " + obj.Properties[key].ToTypeString())
	}
	fmt.Println("}")

	return types
}

type JsonElement interface {
	GetKey() string
	SetKey(key string)
	GetValue() (string, map[string]JsonElement, []JsonElement)
	GetType() string
	IsObject() bool
	ToString(level int) string
	ToTypeString() string
}

type JsonPrimitive struct {
	Key   string
	Value string
	Type  string
}

func (p JsonPrimitive) GetKey() string {
	return p.Key
}

func (p JsonPrimitive) SetKey(key string) {
	p.Key = key
}

func (p JsonPrimitive) GetValue() (string, map[string]JsonElement, []JsonElement) {
	return p.Value, nil, nil
}

func (p JsonPrimitive) GetType() string {
	return p.Type
}

func (p JsonPrimitive) IsObject() bool {
	return false
}

func (p JsonPrimitive) ToString(level int) string {
	padding := strings.Repeat(" ", level*2)

	return padding + p.Key + ": (" + p.GetType() + ") " + p.Value + ",\n"
}

func (p JsonPrimitive) ToTypeString() string {

	return p.Key + ": " + p.GetType() + ",\n"
}

type JsonObject struct {
	Key        string
	Properties map[string]JsonElement
}

func (o JsonObject) GetKey() string {
	return o.Key
}

func (o JsonObject) SetKey(key string) {
	o.Key = key
}

func (o JsonObject) GetValue() (string, map[string]JsonElement, []JsonElement) {
	return "", o.Properties, nil
}

func (o JsonObject) GetType() string {
	if o.Key == "" {
		return "Root"
	}

	res := ""
	if strings.ContainsAny(o.Key, "_-") {
		for i := 0; i < len(o.Key); i++ {
			if o.Key[i] == '_' || o.Key[i] == '-' {
				i += 1
				res += strings.ToUpper(string(o.Key[i]))
			} else if i == 0 {
				res += strings.ToUpper(string(o.Key[i]))
			} else {
				res += string(o.Key[i])
			}
		}
	} else {
		res += strings.ToUpper(string(o.Key[0]))
		res += o.Key[1:len(o.Key)]
	}

	return res
}

func (o JsonObject) IsObject() bool {
	return true
}

func (o JsonObject) ToString(level int) string {
	padding := strings.Repeat(" ", level*2)
	res := padding

	if o.Key != "" {
		res += o.Key + ": "
	}
	res += "(" + o.GetType() + ") "

	res += "{\n"

	for key := range o.Properties {
		res += o.Properties[key].ToString(level + 1)
	}

	res += padding + "},\n"

	return res
}

func (o JsonObject) ToTypeString() string {

	return o.Key + ": " + o.GetType() + ",\n"
}

type JsonArray struct {
	Key        string
	Properties []JsonElement
}

func (a JsonArray) GetKey() string {
	return a.Key
}

func (a JsonArray) SetKey(key string) {
	a.Key = key
}

func (a JsonArray) GetValue() (string, map[string]JsonElement, []JsonElement) {
	return "", nil, a.Properties
}

func (a JsonArray) GetType() string {
	return a.Properties[0].GetType() + "[]"
}

func (a JsonArray) IsObject() bool {
	return false
}

func (a JsonArray) ToString(level int) string {
	padding := strings.Repeat(" ", level*2)
	res := padding

	if a.Key != "" {
		res += a.Key + ": "
	}
	res += "(" + a.GetType() + ") "

	res += "[\n"

	for index := range a.Properties {
		res += a.Properties[index].ToString(level + 1)
	}

	res += padding + "],\n"

	return res
}

func (a JsonArray) ToTypeString() string {
	res := "[\n"

	for key := range a.Properties {
		res += a.Properties[key].ToTypeString()
	}

	res += "],\n"

	return res
}

func parsePrimitive(json string, i int, key string) (JsonPrimitive, int) {
	res := JsonPrimitive{Key: key}
	value := []byte{}
	if unicode.IsNumber(rune(json[i])) {
		res.Type = "number"
		for ; i < len(json) && unicode.IsNumber(rune(json[i])); i++ {
			value = append(value, json[i])
		}
		res.Value = string(value)
		return res, i

	} else if unicode.IsLetter(rune(json[i])) {
		for ; i < len(json) && unicode.IsLetter(rune(json[i])); i++ {
			value = append(value, json[i])
		}
		res.Value = string(value)
		if res.Value == "null" {
			res.Type = "null"
		} else {
			res.Type = "boolean"
		}
		return res, i

	} else if json[i] == '"' {
		res.Type = "string"
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
		i = skipWhiteSpace(json, i)
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
		i = skipWhiteSpace(json, i)
		// panic("")
	}

	return res, i
}

func readKey(json string, i int) (string, int) {
	i = skipWhiteSpace(json, i)
	if json[i] != '"' {
		panic(fmt.Sprintf("Expected start of key at position %d, found %c\nslice: %v", i, json[i], json[i-10:i+10]))
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
