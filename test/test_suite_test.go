package main_test

import (
	"flag"
	"log"
	"os"
	"runtime"
	"testing"

	aero "github.com/aerospike/aerospike-client-go"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	ns  = flag.String("n", "test", "namespace")
	set = flag.String("s", "test", "set")

	host = flag.String("h", "127.0.0.1", "host")
	port = flag.Int("p", 3000, "port")

	user     = flag.String("U", "", "User.")
	password = flag.String("P", "", "Password.")

	recordCount = flag.Int("r", 10000, "number of records")
	nameVariety = flag.Int("v", 100, "number of unique names")

	currentPath = flag.String("lua", "", "Lua Path")

	client *aero.Client
	sqlDB  *sqlx.DB
)

func init_env() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	/****************************************************************************

	Find Current Path

	****************************************************************************/
	var err error
	if len(*currentPath) == 0 {
		*currentPath, err = os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
	}

	/****************************************************************************

	Connect To Aerospike DB

	****************************************************************************/
	client, err = aeroClient(*host, *port, *user, *password, *currentPath)
	if err != nil {
		log.Fatalln("Error connecting to aerospike cluster:", err)
	}

	data := randomRecords(*recordCount, *nameVariety)

	/****************************************************************************

	Setup Aerospike Data

	****************************************************************************/
	if err := aeroSetupDB(client, *currentPath); err != nil {
		log.Fatalln("Error saving test data to Aerospike:", err)
	}

	if err := genAeroData(client, *ns, *set, data); err != nil {
		log.Fatalln("Error saving test data to Aerospike:", err)
	}

	/******************************0*********************************************

	Connect To SQLITE DB

	****************************************************************************/
	sqlDB, err = sqlite3db("file:test.db?cache=shared&mode=memory")
	if err != nil {
		log.Fatalln("Error connecting to sqlite3 db:", err)
	}

	/******************************0*********************************************

	Setup SQLITE Data

	****************************************************************************/
	err = genSqlData(sqlDB, data)
	if err != nil {
		log.Fatalln("Error connecting to sqlite3 db:", err)
	}
}

func TestTest(t *testing.T) {
	init_env()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")

	defer client.Close()
	defer sqlDB.Close()
}
