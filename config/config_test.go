package config_test

import (
	. "github.com/natrim/grainbot/config"
	"io/ioutil"
	"os"
	"testing"
	"bitbucket.org/kardianos/osext"
	"path/filepath"
)

var path string

func init(){
	path, _ = osext.ExecutableFolder() //current bin directory
}

func TestLoad(t *testing.T) {
	//t.Fail()

	if err := ioutil.WriteFile(filepath.Join(path, "test_load.json"), []byte(ExampleConfig().String()), 0664); err != nil {
		t.Log("Nelze vytvořit testovací configurák!")
		t.Log(err)
		t.Fail()
		return
	}

	defer os.Remove(filepath.Join(path, "test_load.json"))

	conf, err := Load("test_load.json")
	if err != nil {
		t.Log("Nelze načíst testovací configurák!")
		t.Log(err)
		t.Fail()
		return
	}

	if conf.String() != ExampleConfig().String() {
		t.Log("Chyba načtení configu!")
		t.Fail()
		return
	}
}

func TestSave(t *testing.T) {
	//t.Fail()

	err := Save(ExampleConfig(), "test_save.json")

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

	if string(data) != ExampleConfig().String() {
		t.Log("Chyba uložení configu!")
		t.Fail()
		return
	}
}

func TestModuleThings(t *testing.T) {
	//t.Fail()

	conf := ExampleConfig()
	conf.Set("test1.string", "stringvalue")
	conf.Set("test2.number", 1337)
	conf.Set("test3.one", "one")
	conf.Set("test3.two", "two")
	conf.Set("test4.one", 1)
	conf.Set("test4.two", 2)
	conf.Set("test5.pony", []string{"RD", "F", "TS", "PP", "AJ", "R"})

	Save(conf, "test_thingies.json")
	conf2, _ := Load("test_thingies.json")

	defer os.Remove("test_thingies.json")

	if conf.String() != conf2.String() {
		t.Log("Chyba configu!")
		t.Fail()
		return
	}

	if conf2.GetInt("test2.number") != 1337 {
		t.Log("Chyba čtení configu 2!")
		t.Log(conf2.String())
		t.Fail()
		return
	}

	if conf2.GetString("test3.one") != "one" {
		t.Log("Chyba čtení configu 3!")
		t.Log(conf2.String())
		t.Fail()
		return
	}

	if conf2.GetInt("test4.two") != 2 {
		t.Log("Chyba čtení configu 4!")
		t.Log(conf2.String())
		t.Fail()
		return
	}

	if conf2.GetStringSlice("test5.pony")[2] != "TS" {
		t.Log("Chyba čtení configu 5!")
		t.Log(conf2.String())
		t.Fail()
		return
	}
}
