package json

import (
	"fmt"
	"strings"
	"unicode"
)

type JsonParser struct {
	json string
	i    int
}

func NewParser(json string) JsonParser {
	return JsonParser{json: json, i: 0}
}

type JsonType = int

const (
	Object  int = 0
	Array       = 1
	Invalid     = -1
)

func (p *JsonParser) GetType() JsonType {
	if p.i != 0 {
		return Invalid
	}

	char := p.json[p.i]
	p.i += 1

	p.skipWhiteSpace()
	switch char {
	case '{':
		return Object

	case '[':
		return Array
	}

	return Invalid
}

func (p *JsonParser) CanParse() bool {
	p.skipWhiteSpace()
	return p.i < len(p.json) && p.json[p.i] != '}'
}

func (p *JsonParser) PeekPropertyName() string {
	start := p.i
	nextProp := p.readKey()
	p.i = start

	return nextProp
}

func (p *JsonParser) ParsePropertyWithEnv(env map[string]string) (*JsonObject, *JsonArray, *JsonPrimitive) {
	propName := p.parseKey()

	switch p.json[p.i] {
	case '{':
		res := JsonObject{}
		res = p.parseObject(propName, env)

		if rune(p.json[p.i]) == ',' {
			p.i += 1
		}
		return &res, nil, nil

	case '[':
		res := JsonArray{}
		res = p.parseArray(propName, env)

		if rune(p.json[p.i]) == ',' {
			p.i += 1
		}
		return nil, &res, nil

	default:
		res := JsonPrimitive{}
		res = p.parsePrimitive(propName, env)

		if rune(p.json[p.i]) == ',' {
			p.i += 1
		}
		return nil, nil, &res
	}

}

func (p *JsonParser) ParseProperty() (*JsonObject, *JsonArray, *JsonPrimitive) {
	return p.ParsePropertyWithEnv(nil)
}

func (p *JsonParser) parseKey() string {
	propName := p.readKey()
	p.skipWhiteSpace()
	if p.json[p.i] != ':' {
		fmt.Println(p.json[p.i:])
		panic(fmt.Sprintf("Malformed json. index: '%d', found '%s', expected ':'.", p.i, string(p.json[p.i])))
	}
	p.i += 1

	p.skipWhiteSpace()

	return propName
}

func (p *JsonParser) ParseJson() (*JsonObject, *JsonArray) {
	p.skipWhiteSpace()
	switch p.json[p.i] {
	case '{':
		res := JsonObject{}
		res = p.parseObject("", nil)
		return &res, nil

	case '[':
		res := JsonArray{}
		res = p.parseArray("", nil)
		return nil, &res
	}

	return nil, nil
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

func (p *JsonParser) parsePrimitive(key string, env map[string]string) JsonPrimitive {
	res := JsonPrimitive{Key: key}
	value := []byte{}
	if unicode.IsNumber(rune(p.json[p.i])) {
		res.Type = "number"
		for ; p.i < len(p.json) && unicode.IsNumber(rune(p.json[p.i])); p.i++ {
			value = append(value, p.json[p.i])
		}
		res.Value = string(value)
		return res

	} else if unicode.IsLetter(rune(p.json[p.i])) {
		for ; p.i < len(p.json) && unicode.IsLetter(rune(p.json[p.i])); p.i++ {
			value = append(value, p.json[p.i])
		}
		res.Value = string(value)
		if res.Value == "null" {
			res.Type = "null"
		} else {
			res.Type = "boolean"
		}
		return res

	} else if p.json[p.i] == '"' {
		res.Type = "string"
		p.i += 1 // skip first "

		for ; p.i < len(p.json) && p.json[p.i] != '"'; p.i++ {
			char := p.json[p.i]
			if env != nil && (rune(char) == '$' && rune(p.json[p.i+1]) == '{') {
				envKey := []byte{}
				p.i += 2
				for ; p.i < len(p.json) && p.json[p.i] != '}'; p.i++ {
					if rune(p.json[p.i]) != '"' {
						envKey = append(envKey, p.json[p.i])
					} else {
						panic("Malformed environment variable, missing closing '}'")
					}
				}
				fmt.Println("envKey:", string(envKey))
				fmt.Println("envVal:", env[string(envKey)])

				value = append(value, []byte(env[string(envKey)])...)
			} else if rune(char) == '\\' {
				p.i += 1
				value = append(value, p.json[p.i])
			} else {
				value = append(value, char)
			}
		}

		res.Value = string(value)

		p.i += 1 // skip last "
		return res

	} else {
		panic("Malformed json in parsePrimitive.")
	}
}

func (p *JsonParser) parseObject(key string, env map[string]string) JsonObject {
	p.i++
	p.skipWhiteSpace()
	res := JsonObject{Key: key, Properties: make(map[string]JsonElement)}

	for p.i < len(p.json) && p.json[p.i] != '}' {
		propName := p.parseKey()

		switch p.json[p.i] {
		case '{':
			res.Properties[propName] = p.parseObject(propName, env)
			break

		case '[':
			res.Properties[propName] = p.parseArray(propName, env)
			break

		default:
			res.Properties[propName] = p.parsePrimitive(propName, env)
			break
		}

		if p.json[p.i] == ',' {
			p.i += 1
		}
		p.skipWhiteSpace()
		// panic("")
	}

	if p.json[p.i] == '}' {
		p.i += 1
	}
	p.skipWhiteSpace()

	return res
}

func (p *JsonParser) parseArray(key string, env map[string]string) JsonArray {
	p.i++
	p.skipWhiteSpace()
	res := JsonArray{Key: key, Properties: []JsonElement{}}

	for p.i < len(p.json) && p.json[p.i] != ']' {
		p.skipWhiteSpace()

		if p.json[p.i] == '{' {
			res.Properties = append(res.Properties, p.parseObject("", env))
		} else {
			res.Properties = append(res.Properties, p.parsePrimitive("", env))
		}

		if p.json[p.i] == ',' {
			p.i += 1
		}
		p.skipWhiteSpace()
		// panic("")
	}
	if p.json[p.i] == ']' {
		p.i += 1
	}
	p.skipWhiteSpace()

	return res
}

func (p *JsonParser) readKey() string {
	p.skipWhiteSpace()
	if p.json[p.i] != '"' {
		panic(fmt.Sprintf("Expected start of key at position %d, found %c\n", p.i, p.json[p.i]))
	}
	p.i += 1

	key := []byte{}
	for p.json[p.i] != '"' {
		key = append(key, p.json[p.i])
		p.i += 1
	}
	p.i += 1

	return string(key)
}

func (p *JsonParser) skipWhiteSpace() {
	for p.i < len(p.json) && unicode.IsSpace(rune(p.json[p.i])) {
		p.i += 1
	}
}
