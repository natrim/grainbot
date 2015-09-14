package config_test

import (
	"github.com/kardianos/osext"
	. "github.com/natrim/grainbot/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var path string

func init() {
	path, _ = osext.ExecutableFolder() //current bin directory
}

func TestLoad(t *testing.T) {
	//t.Fail()

	if err := ioutil.WriteFile(filepath.Join(path, "test_load.json"), []byte(NewExampleConfiguration().String()), 0664); err != nil {
		t.Log("Nelze vytvořit testovací configurák!")
		t.Log(err)
		t.Fail()
		return
	}

	defer os.Remove(filepath.Join(path, "test_load.json"))

	conf, err := LoadConfigFromFile("test_load.json")
	if err != nil {
		t.Log("Nelze načíst testovací configurák!")
		t.Log(err)
		t.Fail()
		return
	}

	if conf.String() != NewExampleConfiguration().String() {
		t.Log("Chyba načtení configu!")
		t.Fail()
		return
	}
}

func TestSave(t *testing.T) {
	//t.Fail()

	err := SaveConfigToFile(NewExampleConfiguration(), "test_save.json")

	if err != nil {
		t.Log("Nelze uložit config!")
		t.Log(err)
		t.Fail()
		return
	}

	defer os.Remove(filepath.Join(path, "test_save.json"))

	data, err := ioutil.ReadFile(filepath.Join(path, "test_save.json"))
	if err != nil {
		t.Log("Nelze načíst uložit config!")
		t.Log(err)
		t.Fail()
		return
	}

	if string(data) != NewExampleConfiguration().String() {
		t.Log("Chyba uložení configu!")
		t.Fail()
		return
	}
}

func TestSelfLoad(t *testing.T) {
	conf := NewExampleConfiguration()
	conf.Server = "pony"
	conf.SaveToFile("test_selfload.json")
	defer os.Remove("test_selfload.json")

	conf = NewConfiguration()

	if conf.Server != "" {
		t.Fail()
	}

	conf.LoadFromFile("test_selfload.json")

	if conf.Server != "pony" {
		t.Log("Failed to load config!")
		t.Fail()
	}
}

func TestModuleThings(t *testing.T) {
	//t.Fail()

	conf := NewExampleConfiguration()
	conf.Set("test1.string", "stringvalue")
	conf.Set("test2.number", 1337)
	conf.Set("test3.one", "one")
	conf.Set("test3.two", "two")
	conf.Set("test4.one", 1)
	conf.Set("test4.two", 2)
	conf.Set("test5.pony", []string{"RD", "F", "TS", "PP", "AJ", "R"})
	conf.Set("test6.on", true)
	conf.Set("test7.one", 6.1337)
	conf.Set("test7.two", "6.1337")
	conf.Set("test7.three", "6")
	conf.Set("test8.count", []int{1, 2, 3, 4, 5, 6})
	conf.Set("test9.time1", "2006-01-02 15:04:05 -0700")
	now := time.Now()
	conf.Set("test9.time2", now)

	SaveConfigToFile(conf, "test_thingies.json")
	conf2, _ := LoadConfigFromFile("test_thingies.json")

	defer os.Remove("test_thingies.json")

	if conf.String() != conf2.String() {
		t.Log("Chyba configu!")
		t.Log(conf.String())
		t.Log(conf2.String())
		t.Fail()
		return
	}

	if r, err := conf2.GetInt("test2.number"); r != 1337 {
		t.Log("Chyba čtení configu 2!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetString("test3.one"); r != "one" {
		t.Log("Chyba čtení configu 3!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetInt("test4.two"); r != 2 {
		t.Log("Chyba čtení configu 4!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetStringSlice("test5.pony"); r[2] != "TS" {
		t.Log("Chyba čtení configu 5!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetBool("test6.on"); r != true {
		t.Log(err)
		t.Log("Chyba čtení configu 6!")
		t.Fail()
		return
	}

	if r, err := conf2.GetFloat("test7.one"); r != 6.1337 {
		t.Log("Chyba čtení configu 7!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetFloat("test7.two"); r != 6.1337 {
		t.Log("Chyba čtení configu 7!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetFloat("test7.three"); r != 6 {
		t.Log("Chyba čtení configu 7 vint!")
		t.Log(err)
		t.Fail()
		return
	}

	// fails
	if _, r := conf2.GetInt("test.error"); r == nil {
		t.Log("Chyba test GetInt error!")
		t.Fail()
		return
	}
	if _, r := conf2.GetFloat("test.error"); r == nil {
		t.Log("Chyba test GetFloat error!")
		t.Fail()
		return
	}

	if _, r := conf2.GetString("test.error"); r == nil {
		t.Log("Chyba test GetString error!")
		t.Fail()
		return
	}

	if _, r := conf2.GetStringSlice("test.error"); r == nil {
		t.Log("Chyba test GetStringSlice error!")
		t.Fail()
		return
	}

	if _, r := conf2.GetBool("test.error"); r == nil {
		t.Log("Chyba test GetBool error!")
		t.Fail()
		return
	}

	if r, err := conf2.GetIntSlice("test8.count"); r[2] != 3 {
		t.Log("Chyba čtení configu 8!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetTime("test9.time1"); r.Format("2006-01-02 15:04:05 -0700") != "2006-01-02 15:04:05 -0700" {
		t.Log("Chyba čtení configu 9!")
		t.Log(err)
		t.Fail()
		return
	}

	if r, err := conf2.GetTime("test9.time2"); r.String() != now.String() {
		t.Log("Chyba čtení configu 9 - now!")
		t.Log(err)
		t.Fail()
		return
	}
}
