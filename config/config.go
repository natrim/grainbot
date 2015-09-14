package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kardianos/osext"
	"github.com/spf13/cast"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Configuration struct
type Configuration struct {
	filepath string
	Server   string
	SSL      bool
	Nick     string
	UserName string
	RealName string

	Owner     string
	UpdateUrl string

	Modules map[string]interface{}

	sync.RWMutex
}

// NewConfiguration returns empty Configurations instance
func NewConfiguration() *Configuration {
	return &Configuration{}
}

// LoadFromFile loads configuration from file
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

// Load loads configuration from default file
func (conf *Configuration) Load() error {
	return conf.LoadFromFile("")
}

// LoadConfigFromFile returns new Configuration instance loaded from file
func LoadConfigFromFile(file string) (*Configuration, error) {
	conf := &Configuration{}
	return conf, conf.LoadFromFile(file)
}

// LoadConfig returns new Configuration instance loaded from default file
func LoadConfig() (*Configuration, error) {
	conf := &Configuration{}
	return conf, conf.LoadFromFile("")
}

// SaveToFile saves configuration to file
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

// SaveConfigToFile saves configuration to file
func SaveConfigToFile(conf *Configuration, file string) error {
	return conf.SaveToFile(file)
}

// Save saves configuration to default file
func (conf *Configuration) Save() error {
	return conf.SaveToFile("")
}

// SaveConfig saves Configuration to default file
func SaveConfig(conf *Configuration) error {
	return conf.SaveToFile("")
}

// Set module config value
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
	case []string, []int:
	case int, int8, int16, int32, int64:
	case float32, float64:
	case nil:
	case bool:
	case time.Time:
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

// Get module config
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

// GetString module config value
func (conf *Configuration) GetString(key string) (string, error) {
	ret, err := conf.Get(key)

	if err != nil {
		return "", err
	}

	return cast.ToStringE(ret)
}

// GetInt module config value
func (conf *Configuration) GetInt(key string) (int, error) {
	ret, err := conf.Get(key)

	if err != nil {
		return 0, err
	}

	return cast.ToIntE(ret)
}

// GetFloat module config value
func (conf *Configuration) GetFloat(key string) (float64, error) {
	ret, err := conf.Get(key)

	if err != nil {
		return 0.0, err
	}

	return cast.ToFloat64E(ret)
}

// GetBool module config value
func (conf *Configuration) GetBool(key string) (bool, error) {
	ret, err := conf.Get(key)

	if err != nil {
		return false, err
	}

	return cast.ToBoolE(ret)
}

// GetStringSlice module config value
func (conf *Configuration) GetStringSlice(key string) ([]string, error) {
	ret, err := conf.Get(key)

	var a []string

	if err != nil {
		return a, err
	}

	return cast.ToStringSliceE(ret)
}

// GetIntSlice module config value
func (conf *Configuration) GetIntSlice(key string) ([]int, error) {
	ret, err := conf.Get(key)

	var a []int

	if err != nil {
		return a, err
	}

	return cast.ToIntSliceE(ret)
}

// GetTime module config value
func (conf *Configuration) GetTime(key string) (time.Time, error) {
	ret, err := conf.Get(key)

	if err != nil {
		return time.Time{}, err
	}

	return cast.ToTimeE(ret)
}

// String returns configuration as json encoded string
func (conf *Configuration) String() string {
	conf.RLock()
	defer conf.RUnlock()

	cbuf, err := json.MarshalIndent(conf, "", "    ")

	if err != nil {
		return "{\"error\": \" " + err.Error() + " \"}"
	}

	return string(cbuf)
}

// LoadExampleConfig loads example config
func (conf *Configuration) LoadExampleConfig() {
	conf.Server = "irc.freenode.net:6667"
	conf.Owner = "Natrim"
	conf.UpdateUrl = "http://natrim.cz/uploads/grainbot_linux"
	conf.Modules = map[string]interface{}{"autojoin": map[string]interface{}{"channels": []string{"#pony"}}}
}

// NewExampleConfiguration returns new Configuration instance with example values
func NewExampleConfiguration() *Configuration {
	conf := NewConfiguration()
	conf.LoadExampleConfig()
	return conf
}
