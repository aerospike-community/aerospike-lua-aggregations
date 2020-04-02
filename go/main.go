package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	aero "github.com/aerospike/aerospike-client-go"
)

var (
	host        = flag.String("h", "127.0.0.1", "host")
	port        = flag.Int("p", 3000, "port")
	user        = flag.String("U", "", "User.")
	password    = flag.String("P", "", "Password.")
	currentPath = flag.String("dir", "", "Lua Path")

	// currentPath string
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	clientPolicy := aero.NewClientPolicy()
	if *user != "" {
		clientPolicy.User = *user
		clientPolicy.Password = *password
	}

	client, err := aero.NewClientWithPolicyAndHost(clientPolicy, aero.NewHost(*host, *port))
	if err != nil {
		log.Fatalln("Error connecting to the DB:", err)
	}
	defer client.Close()

	if len(*currentPath) == 0 {
		*currentPath, err = os.Getwd()
		if err != nil {
			log.Println(err)
		}
	}

	if err := setupDB(client); err != nil {
		log.Fatalln("Error registering the UDF:", err)
	}

	aero.SetLuaPath(filepath.Clean(*currentPath) + "/")
	if err := queryAggregate(client, "test", "test"); err != nil {
		log.Fatalln(err)
	}
}

func queryAggregate(client *aero.Client, nsName, setName string) error {
	stm := aero.NewStatement(nsName, setName)

	functionArgsMap := map[string]interface{}{
		"fields": map[string]interface{}{
			"name":              "name",
			"doesnt_exist":      "doesnt_exist",
			"max(doesnt_exist)": map[string]string{"func": "max", "expr": "rec['doesnt_exist']"},
			"max(age)":          map[string]string{"func": "max", "expr": "rec['age']"},
			"count(age)":        map[string]string{"func": "count", "expr": "rec['age'] ~= nil and 1"},
			"min(age)":          map[string]string{"func": "min", "expr": "rec['age']"},
			"sum(age*salary)":   map[string]string{"func": "sum", "expr": "(rec['age']  or 0) * (rec['salary'] or 0)"},
			"sum(age)":          map[string]string{"func": "sum", "expr": "rec['age']"},
		},
		"filter": "rec['age'] ~= nil and rec['age'] > 25",
		"group_by_fields": []string{
			"name",
			"lastname",
		},
	}

	recordset, err := client.QueryAggregate(nil, stm, "aggAPI", "select_agg_records", aero.NewValue(functionArgsMap))
	defer recordset.Close()

	if err != nil {
		return err
	}

	for result := range recordset.Results() {
		if result.Err != nil {
			return result.Err
		}

		// pp.Println(result.Record.Bins["SUCCESS"])
		fmt.Println("DONE!", len(result.Record.Bins["SUCCESS"].(map[interface{}]interface{})))
	}

	return nil
}

func setupDB(client *aero.Client) error {
	fileName := filepath.Join(*currentPath, "aggAPI.lua")
	luaFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	regTask, err := client.RegisterUDF(nil, luaFile, "aggAPI.lua", aero.LUA)
	if err != nil {
		return err
	}

	// wait until UDF is created
	err = <-regTask.OnComplete()
	if err != nil {
		return err
	}

	return nil
}
