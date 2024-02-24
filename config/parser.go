package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"tobiasthedanish/thetypeagen/json"
)

type Config struct {
	Output    string
	Endpoints []EndpointConfig
}

type EndpointConfig struct {
	RootType string
	Url      string
	Options  RequestOptions
}

type RequestOptions struct {
	Method  string
	Headers http.Header
	Body    io.Reader
}

func (c *Config) ToString() string {
	res := "Config: [\n"

	for _, epc := range c.Endpoints {
		res += "  {\n"
		res += "    RootType: " + epc.RootType + ",\n"
		res += "    Url: " + epc.Url + ",\n"
		res += "    Options: {\n"
		res += "      Method: " + epc.Options.Method + ",\n"
		res += "      Headers: " + strconv.Itoa(len(epc.Options.Headers)) + ",\n"
		res += "  },\n"
	}

	res += "]\n"

	return res
}

func Fetch(url string, options RequestOptions) (*http.Response, error) {
	req, err := http.NewRequest(options.Method, url, options.Body)
	if err != nil {
		return nil, err
	}

	req.Header = options.Headers
	return http.DefaultClient.Do(req)
}

func GetBodyAsString(res *http.Response) (string, error) {
	defer res.Body.Close()

	body := ""
	scanner := bufio.NewScanner(res.Body)
	for i := 0; scanner.Scan(); i++ {
		body = body + scanner.Text()
	}

	return body, nil
}

func readFile(name string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(wd + name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	//
	// bytes, err := os.ReadFile(f.Name())
	// if err != nil {
	// 	return nil, err
	// }

	reader := bufio.NewReader(f)

	bytes := make([]byte, reader.Size())
	_, err = reader.Read(bytes)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return bytes, nil
}

func ParseConfig() (*Config, error) {
	bytes, err := readFile("/.typeagenrc.json")
	if err != nil {
		return nil, err
	}

	configStr := string(bytes)
	parser := json.NewParser(configStr)

	if parser.GetType() != json.Object {
		return nil, errors.New("Config should always be a json object")
	}

	var (
		environment map[string]string
		outfile     string
		endpoints   []json.JsonElement
	)
	for parser.CanParse() {
		next := parser.PeekPropertyName()

		switch next {
		case "envFile":
			_, _, env := parser.ParseProperty()
			if env == nil {
				return nil, errors.New("Property 'envFile' is expected to be a string")
			}
			environment = parseEnvFile(env.Value)
			break

		case "outFile":
			_, _, out := parser.ParsePropertyWithEnv(environment)
			if out == nil {
				return nil, errors.New("Property 'outFile' is expected to be a string")
			}
			outfile = out.Value
			break

		case "endpoints":
			_, ep, _ := parser.ParsePropertyWithEnv(environment)
			if ep == nil {
				return nil, errors.New("Property 'endpoints' is expected to be an array")
			}
			endpoints = ep.Properties
		}
	}

	config := Config{
		Output:    outfile,
		Endpoints: []EndpointConfig{},
	}

	for _, ep := range endpoints {
		if !ep.IsObject() {
			return nil, errors.New("Expected 'endpoints' to be an array of objects.")
		}
		epObj := ep.(json.JsonObject)
		epProps := epObj.Properties
		rootType, err := getRequiredPrimitiveProp("rootType", epProps)
		if err != nil {
			return nil, err
		}

		url, err := getRequiredPrimitiveProp("url", epProps)
		if err != nil {
			return nil, err
		}

		method, err := getRequiredPrimitiveProp("method", epProps)
		if err != nil {
			return nil, err
		}

		epConfig := EndpointConfig{
			RootType: rootType,
			Url:      url,
			Options: RequestOptions{
				Method:  method,
				Headers: make(http.Header),
			},
		}

		headers := getOptionalObjectProp("headers", epProps)

		if headers != nil {
			for key := range headers {
				val, err := getRequiredPrimitiveProp(key, headers)
				if err != nil {
					return nil, err
				}
				epConfig.Options.Headers.Add(key, val)
			}
		}

		config.Endpoints = append(config.Endpoints, epConfig)
	}

	return &config, nil
}

func getOptionalObjectProp(propName string, props map[string]json.JsonElement) map[string]json.JsonElement {
	if props[propName] != nil {
		_, val, _ := props[propName].GetValue()

		return val
	} else {
		return nil
	}
}

func getRequiredPrimitiveProp(propName string, props map[string]json.JsonElement) (string, error) {
	prop, _, _ := props[propName].GetValue()
	if prop == "" {
		return "", errors.New(propName + " is a required prop for endpoint configuration.")
	}

	return prop, nil
}

func getOptionalPrimitiveProp(propName string, props map[string]json.JsonElement) string {
	if props[propName] != nil {
		val, _, _ := props[propName].GetValue()

		return val
	} else {
		return ""
	}
}

func parseEnvFile(filename string) map[string]string {
	envVars := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.Trim(line, " \n\r\t")) == 0 {
			continue
		}

		keyVal := strings.Split(line, "=")
		key := strings.Trim(keyVal[0], " \n\r\t")
		val := strings.Trim(keyVal[1], " \n\r\t")

		envVars[key] = val
	}

	return envVars
}
