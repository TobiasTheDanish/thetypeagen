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

func ParseConfig() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(wd + "/.typeagenrc.json")
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

	configStr := string(bytes)
	parser := json.NewParser(configStr)

	obj, arr := parser.ParseJson()
	if arr != nil {
		return nil, errors.New("Config is expected to start with a JSON object, found JSON array")
	}

	props := obj.Properties

	outFile := getOptionalPrimitiveProp("outFile", props)
	envFile := getOptionalPrimitiveProp("envFile", props)

	envVars := parseEnvFile(envFile)

	for key, val := range envVars {
		fmt.Println(fmt.Sprintf("%s=%s", key, val))
	}

	config := Config{
		Output:    outFile,
		Endpoints: []EndpointConfig{},
	}

	endpoints := props["endpoints"]
	_, _, arrVal := endpoints.GetValue()
	if arrVal == nil {
		return nil, errors.New("Expected 'endpoints' to be an array.")
	}

	for _, ep := range arrVal {
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
