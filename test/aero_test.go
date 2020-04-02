package main_test

import (
	"io/ioutil"
	"log"
	"path/filepath"

	aero "github.com/aerospike/aerospike-client-go"
)

func aeroClient(host string, port int, user, password, currentPath string) (*aero.Client, error) {
	/****************************************************************************

	Setup Aerospike Client

	****************************************************************************/
	clientPolicy := aero.NewClientPolicy()
	if user != "" {
		clientPolicy.User = user
		clientPolicy.Password = password
	}

	client, err := aero.NewClientWithPolicyAndHost(clientPolicy, aero.NewHost(host, port))
	if err != nil {
		return nil, err
	}

	/****************************************************************************

	Setup Aerospike Lua Path

	****************************************************************************/
	aero.SetLuaPath(filepath.Clean(currentPath) + "/")

	return client, nil
}

func toInt64(v interface{}) int64 {
	switch v := v.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		panic("Not possible")
	}
}

func aeroQuery(client *aero.Client, nsName, setName string, payload map[string]interface{}) ([]map[string]interface{}, error) {
	stm := aero.NewStatement(nsName, setName)

	recordset, err := client.QueryAggregate(nil, stm, "aggAPI", "select_agg_records", aero.NewValue(payload))
	defer recordset.Close()

	if err != nil {
		return nil, err
	}

	res := []map[string]interface{}{}
	for result := range recordset.Results() {
		if result.Err != nil {
			return nil, result.Err
		}

		v := result.Record.Bins["SUCCESS"].(map[interface{}]interface{})
		for _, prec := range v {
			rres := map[string]interface{}{}
			for k, v := range prec.(map[interface{}]interface{}) {
				if i, ok := v.(float64); ok {
					rres[k.(string)] = toInt64(i)
				} else if i, ok := v.(int); ok {
					rres[k.(string)] = toInt64(i)
				} else if s, ok := v.(string); ok {
					rres[k.(string)] = s
				}
			}

			res = append(res, rres)
		}
	}

	return res, nil
}

func genAeroData(client *aero.Client, ns, set string, data []map[string]interface{}) error {
	err := client.Truncate(nil, ns, set, nil)
	if err != nil {
		return err
	}

	for i := range data {
		key, _ := aero.NewKey(ns, set, data[i]["id"])
		if err := client.Put(nil, key, data[i]); err != nil {
			return err
		}

		if i > 0 && i%1000 == 0 {
			log.Println("Aerospike progress:", i)
		}
	}

	log.Println("Aerospike completed successfully...", len(data))

	return nil
}

func aeroSetupDB(client *aero.Client, currentPath string) error {
	fileName := filepath.Join(currentPath, "aggAPI.lua")
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
