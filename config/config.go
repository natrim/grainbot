package config

import (
	"bitbucket.org/kardianos/osext"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Configuration struct {
	HostName string
	Port     int
	SSL      bool
	Nick     string
	UserName string
	RealName string
	Modules  map[string]interface{}
}

func New() *Configuration {
	return &Configuration{}
}

func (conf *Configuration) Load(file string) error {
	conf, err := Load(file)
	return err
}

func Load(file string) (*Configuration, error) {
	var pathToConfig string
	var err error
	var loaded bool

	if file != "" {
		if filepath.IsAbs(file) {
			pathToConfig = file
		} else {
			var filename string
			filename, err = filepath.Abs(file)
			if err == nil {
				if _, err := os.Stat(filename); !os.IsNotExist(err) {
					pathToConfig = filename
				}
			}
		}
	}

	if pathToConfig == "" {
		var path string
		path, err = osext.ExecutableFolder() //current bin directory
		if err == nil {
			var filename string
			if file == "" {
				filename = filepath.Join(path, "config.json")
			} else {
				filename = filepath.Join(path, file)
			}
			if _, err := os.Stat(filename); !os.IsNotExist(err) {
				pathToConfig = filename
			}
		}
	}

	var data *Configuration
	if pathToConfig != "" {
		var buff []byte
		buff, err = ioutil.ReadFile(pathToConfig)

		if err == nil {
			data = &Configuration{}
			err = json.Unmarshal(buff, data)
			if err == nil {
				loaded = true
			}
		}
	}

	if !loaded {
		if err != nil {
			return nil, errors.New("Cannot load config file! " + err.Error())
		} else {
			return nil, errors.New("Cannot load config file!")
		}
	}

	return data, nil
}

func (conf *Configuration) Save(file string) error {
	return Save(conf, file)
}

func Save(conf *Configuration, file string) error {
	if conf == nil {
		return errors.New("I need valid Configuration to save!")
	}

	var filename string
	var err error

	if file == "" {
		var path string
		path, err = osext.ExecutableFolder() //current bin directory
		if err != nil {
			filename = filepath.Join(path, "config.json")
		} else {
			filename = "config.json"
		}
	} else if !filepath.IsAbs(file) {
		var path string
		path, err = osext.ExecutableFolder() //current bin directory
		if err != nil {
			filename = filepath.Join(path, file)
		} else {
			filename = file
		}
	}

	if filename != "" {
		var cbuf []byte
		cbuf, err = json.Marshal(conf)
		if err == nil {
			err = ioutil.WriteFile(filename, cbuf, 0644)
		}
	}

	if err != nil {
		return errors.New("Cannot save config file! " + err.Error())
	}

	return nil
}

func (conf *Configuration) Set(key string, value interface{}) error {
	key = strings.ToLower(key)

	keys := strings.Split(key, ".")
	if len(keys) <= 1 {
		return errors.New("You need to specify from what module you want to get data! \"(syntax: module.key)\"")
	}

	switch value.(type) {
	case string:
	case []string:
	case int:
	case nil:
	case bool:
	default:
		return errors.New(fmt.Sprintf("%s %T", "Unsupported type! ", value))
	}

	if _, err := conf.Modules[keys[0]]; err {
		conf.Modules[keys[0]].(map[string]interface{})[keys[1]] = value
	} else {
		conf.Modules[keys[0]] = map[string]interface{}{keys[1]: value}
	}

	return nil
}

func (conf *Configuration) Get(key string) interface{} {
	key = strings.ToLower(key)

	keys := strings.Split(key, ".")

	if len(keys) <= 1 {
		return errors.New("You need to specify from what module you want to get data! \"(syntax: module.key)\"")
	}

	ret := conf.Modules[keys[0]]
	if ret == nil {
		return nil
	}

	return ret.(map[string]interface{})[keys[1]]
}

func (conf *Configuration) GetString(key string) string {
	ret := conf.Get(key)

	switch s := ret.(type) {
	case string:
		return s
	case float64:
		return strconv.FormatFloat(ret.(float64), 'f', -1, 64)
	case int:
		return strconv.FormatInt(int64(ret.(int)), 10)
	case []byte:
		return string(s)
	case nil:
		return ""
	default:
		return ""
	}
}

func (conf *Configuration) GetInt(key string) int {
	ret := conf.Get(key)

	switch s := ret.(type) {
	case int:
		return s
	case int64:
		return int(s)
	case int32:
		return int(s)
	case int16:
		return int(s)
	case int8:
		return int(s)
	case string:
		v, err := strconv.ParseInt(s, 0, 0)
		if err == nil {
			return int(v)
		} else {
			return 0
		}
	case float64:
		return int(s)
	case bool:
		if bool(s) {
			return 1
		} else {
			return 0
		}
	case nil:
		return 0
	default:
		return 0
	}
}

func (conf *Configuration) GetBool(key string) bool {
	ret := conf.Get(key)

	switch b := ret.(type) {
	case bool:
		return b
	case nil:
		return false
	case int:
		if ret.(int) > 0 {
			return true
		}
		return false
	case string:
		ret1, err := strconv.ParseBool(ret.(string))
		if err != nil {
			return false
		}
		return ret1
	default:
		return false
	}
}

func (conf *Configuration) GetStringSlice(key string) []string {
	ret := conf.Get(key)

	var a []string

	switch v := ret.(type) {
	case []interface{}:
		for _, u := range v {

			var w string

			switch s := u.(type) {
			case string:
				w = s
			case float64:
				w = strconv.FormatFloat(u.(float64), 'f', -1, 64)
			case int:
				w = strconv.FormatInt(int64(u.(int)), 10)
			case []byte:
				w = string(s)
			case nil:
				w = ""
			default:
				w = ""
			}

			a = append(a, w)
		}
		return a
	case []string:
		return v
	default:
		return a
	}
}

func (conf *Configuration) String() string {
	cbuf, _ := json.Marshal(conf)
	return string(cbuf)
}

func ExampleConfig() *Configuration {
	return &Configuration{HostName: "walter.ecstasy.cz", Modules: map[string]interface{}{"autojoin": []string{"#pony"}}}
}
