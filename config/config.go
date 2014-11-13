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
	"sync"
)

type Configuration struct {
	filepath string
	HostName string
	Port     int
	SSL      bool
	Nick     string
	UserName string
	RealName string
	Modules  map[string]interface{}

	sync.RWMutex
}

func NewConfiguration() *Configuration {
	return &Configuration{}
}

func (conf *Configuration) LoadFromFile(file string) error {
	var pathToConfig string
	var err error
	var loaded bool

	conf.Lock()
	defer conf.Unlock()

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

	if pathToConfig != "" {
		var buff []byte
		buff, err = ioutil.ReadFile(pathToConfig)

		if err == nil {
			err = json.Unmarshal(buff, conf)
			if err == nil {
				loaded = true
				conf.filepath = pathToConfig
			}
		}
	}

	if !loaded {
		if err != nil {
			return errors.New("Cannot load config file! " + err.Error())
		} else {
			return errors.New("Cannot load config file!")
		}
	}

	return nil
}

func (conf *Configuration) Load() error {
	return conf.LoadFromFile("")
}

func LoadConfigFromFile(file string) (*Configuration, error) {
	conf := &Configuration{}
	return conf, conf.LoadFromFile(file)
}

func LoadConfig() (*Configuration, error) {
	conf := &Configuration{}
	return conf, conf.LoadFromFile("")
}

func (conf *Configuration) SaveToFile(file string) error {
	if conf == nil {
		return errors.New("I need valid Configuration to save!")
	}

	var filename string
	var err error

	conf.Lock()
	defer conf.Unlock()

	if file == "" {
		if conf.filepath != "" {
			filename = conf.filepath
		} else {
			var path string
			path, err = osext.ExecutableFolder() //current bin directory
			if err == nil {
				filename = filepath.Join(path, "config.json")
			} else {
				filename = "config.json"
			}
		}
	} else if !filepath.IsAbs(file) {
		var path string
		path, err = osext.ExecutableFolder() //current bin directory
		if err == nil {
			filename = filepath.Join(path, file)
		} else {
			filename = file
		}
	}

	if filename != "" {
		var cbuf []byte
		cbuf, err = json.MarshalIndent(conf, "", "    ")
		if err == nil {
			err = ioutil.WriteFile(filename, cbuf, 0644)
		}
	}

	if err != nil {
		return errors.New("Cannot save config file! " + err.Error())
	}

	return nil
}

func SaveConfigToFile(conf *Configuration, file string) error {
	return conf.SaveToFile(file)
}

func (conf *Configuration) Save() error {
	return conf.SaveToFile("")
}

func SaveConfig(conf *Configuration) error {
	return conf.SaveToFile("")
}

func (conf *Configuration) Set(key string, value interface{}) error {
	conf.Lock()
	defer conf.Unlock()

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

func (conf *Configuration) Get(key string) (interface{}, error) {
	conf.RLock()
	defer conf.RUnlock()

	key = strings.ToLower(key)

	keys := strings.Split(key, ".")

	if len(keys) <= 1 {
		return nil, errors.New("You need to specify from what module you want to get data! \"(syntax: module.key)\"")
	}

	ret, ok := conf.Modules[keys[0]]
	if !ok {
		return nil, errors.New("Module cofiguration not found!")
	}

	return ret.(map[string]interface{})[keys[1]], nil
}

func (conf *Configuration) GetString(key string) string {
	ret, _ := conf.Get(key)

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
	ret, _ := conf.Get(key)

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
	ret, _ := conf.Get(key)

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
	ret, _ := conf.Get(key)

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
	conf.RLock()
	defer conf.RUnlock()

	cbuf, err := json.MarshalIndent(conf, "", "    ")

	if err != nil {
		return "{\"error\": \" " + err.Error() + " \"}"
	}

	return string(cbuf)
}

func ExampleConfig() *Configuration {
	return &Configuration{HostName: "irc.deltaanime.net", Modules: map[string]interface{}{"autojoin": map[string]interface{}{"channels": []string{"#pony"}}}}
}
