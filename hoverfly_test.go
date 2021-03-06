package hoverfly

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestGetNewHoverflyCheckConfig(t *testing.T) {

	cfg := InitSettings()
	cfg.DatabaseName = "testing2.db"
	// getting boltDB
	db := GetDB(cfg.DatabaseName)
	cache := NewBoltDBCache(db, []byte(RequestsBucketName))
	defer cache.CloseDB()

	_, dbClient := GetNewHoverfly(cfg, cache)

	expect(t, dbClient.Cfg, cfg)

	// deleting this database
	os.Remove(cfg.DatabaseName)
}

func TestProcessCaptureRequest(t *testing.T) {
	server, dbClient := testTools(201, `{'message': 'here'}`)
	defer server.Close()
	defer dbClient.Cache.DeleteData()

	r, err := http.NewRequest("GET", "http://somehost.com", nil)
	expect(t, err, nil)

	dbClient.Cfg.SetMode("capture")

	req, resp := dbClient.processRequest(r)

	refute(t, req, nil)
	refute(t, resp, nil)
	expect(t, resp.StatusCode, 201)
}

func TestProcessVirtualizeRequest(t *testing.T) {
	server, dbClient := testTools(201, `{'message': 'here'}`)
	defer server.Close()
	defer dbClient.Cache.DeleteData()

	r, err := http.NewRequest("GET", "http://somehost.com", nil)
	expect(t, err, nil)

	// capturing
	dbClient.Cfg.SetMode("capture")
	req, resp := dbClient.processRequest(r)

	refute(t, req, nil)
	refute(t, resp, nil)
	expect(t, resp.StatusCode, 201)

	// virtualizing
	dbClient.Cfg.SetMode("virtualize")
	newReq, newResp := dbClient.processRequest(r)

	refute(t, newReq, nil)
	refute(t, newResp, nil)
	expect(t, newResp.StatusCode, 201)
}

func TestProcessSynthesizeRequest(t *testing.T) {
	server, dbClient := testTools(201, `{'message': 'here'}`)
	defer server.Close()
	defer dbClient.Cache.DeleteData()

	// getting reflect middleware
	dbClient.Cfg.Middleware = "./examples/middleware/reflect_body/reflect_body.py"

	bodyBytes := []byte("request_body_here")

	r, err := http.NewRequest("GET", "http://somehost.com", ioutil.NopCloser(bytes.NewBuffer(bodyBytes)))
	expect(t, err, nil)

	dbClient.Cfg.SetMode("synthesize")
	newReq, newResp := dbClient.processRequest(r)

	refute(t, newReq, nil)
	refute(t, newResp, nil)
	expect(t, newResp.StatusCode, 200)
	b, err := ioutil.ReadAll(newResp.Body)
	expect(t, err, nil)
	expect(t, string(b), string(bodyBytes))
}

func TestProcessModifyRequest(t *testing.T) {
	server, dbClient := testTools(201, `{'message': 'here'}`)
	defer server.Close()

	// getting reflect middleware
	dbClient.Cfg.Middleware = "./examples/middleware/modify_request/modify_request.py"

	r, err := http.NewRequest("POST", "http://somehost.com", nil)
	expect(t, err, nil)

	dbClient.Cfg.SetMode("modify")
	newReq, newResp := dbClient.processRequest(r)

	refute(t, newReq, nil)
	refute(t, newResp, nil)

	expect(t, newResp.StatusCode, 202)
}
