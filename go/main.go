package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	aero "github.com/aerospike/aerospike-client-go"
	"github.com/k0kubun/pp"
)

var (
	host     = flag.String("h", "127.0.0.1", "host")
	port     = flag.Int("p", 3000, "port")
	user     = flag.String("U", "", "User.")
	password = flag.String("P", "", "Password.")

	currentPath string
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

	currentPath, err = os.Getwd()
	if err != nil {
		log.Println(err)
	}

	if err := setupDB(client); err != nil {
		log.Fatalln("Error registering the UDF:", err)
	}

	aero.SetLuaPath(currentPath + "/")
	if err := queryAggregate(client, "test", "test"); err != nil {
		log.Fatalln(err)
	}
}

func queryAggregate(client *aero.Client, nsName, setName string) error {
	stm := aero.NewStatement(nsName, setName)

	functionArgsMap := map[string]interface{}{
		"raw_fields": map[string]string{
			"name": "name",
		},
		"aggregate_fields": map[string]string{
			"sum(age)":   "sum",
			"min(age)":   "min",
			"max(age)":   "max",
			"count(age)": "count",
		},
		"group_by_fields": []string{
			"name",
		},
		"filter": "if rec['age'] ~= nil and rec['age'] >5  then select_rec = true end",
		"field_aliases": map[string]string{
			"max(age)":        "result =  rec['age'] ~= nil and rec['age']",
			"count(age)":      "result = ( rec['age'] ) ~= nil and 1",
			"min(age)":        "result =  rec['age'] ~= nil and rec['age']",
			"sum(age*salary)": "result =  (rec['age']  or 0) * (rec['salary'] or 0)",
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

		pp.Println(result.Record.Bins["SUCCESS"])
	}

	return nil
}

func setupDB(client *aero.Client) error {
	luaFile, err := ioutil.ReadFile(currentPath + "/aggAPI.lua")
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
