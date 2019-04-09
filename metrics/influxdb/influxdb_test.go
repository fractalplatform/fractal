package influxdb

import (
	"fmt"
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/fractalplatform/fractal/metrics"
	client "github.com/influxdata/influxdb1-client"
)

const (
	dburl     = "http://localhost:8086"
	testdb    = "testmetrics"
	username  = ""
	password  = ""
	namespace = "test/"
	prefix    = "test"
	table     = namespace + prefix + ".timer"
)

func TestWrite(t *testing.T) {
	go InfluxDBWithTags(metrics.DefaultRegistry, 1*time.Second, dburl, testdb, "", "", namespace, make(map[string]string))
	tm := metrics.NewRegisteredTimer(prefix, nil)
	for i := 0; i < 5; i++ {
		tm.Update(100 * time.Second)
	}
	time.Sleep(time.Duration(10) * time.Second)
}

func TestQuery(t *testing.T) {
	host, err := url.Parse(dburl)
	if err != nil {
		log.Fatal(err)
	}
	con, err := client.NewClient(client.Config{URL: *host})
	if err != nil {
		log.Fatal(err)
	}

	q := client.Query{
		Command:  fmt.Sprintf(`select * from "%s"`, table),
		Database: testdb,
	}
	if response, err := con.Query(q); err == nil && response.Error() == nil {
		log.Println(response.Results)
	}
}
