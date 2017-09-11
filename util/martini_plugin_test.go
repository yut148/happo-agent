package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/heartbeatsjp/happo-agent/halib"

	"github.com/go-martini/martini"
	"github.com/stretchr/testify/assert"
)

func TestACL0(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{"FAIL"}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "192.168.0.1", halib.DefaultAgentPort)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, http.StatusServiceUnavailable, res.Code)
}

func TestACL1(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "192.168.0.1", halib.DefaultAgentPort)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, http.StatusForbidden, res.Code)
}

func TestACL2(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, halib.DefaultAgentPort)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, http.StatusOK, res.Code)
	assert.EqualValues(t, bodyStr, res.Body.String())
}

func TestACL3(t *testing.T) {
	const IP = "12.12.12.12"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{IP}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", "127.0.0.1", halib.DefaultAgentPort)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, http.StatusOK, res.Code)
	assert.EqualValues(t, bodyStr, res.Body.String())
}

func TestACL4(t *testing.T) {
	const IP = "12.12.12.12"
	const ipScope = "12.12.12.0/24"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{ipScope}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, halib.DefaultAgentPort)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, http.StatusOK, res.Code)
	assert.EqualValues(t, bodyStr, res.Body.String())
}

func TestACL5(t *testing.T) {
	const IP = "12.12.12.12"
	const ipScope = "192.168.0.0/24"
	const bodyStr = "success"

	m := martini.Classic()
	m.Use(ACL([]string{ipScope}))

	m.Get(("/test"), func() string {
		return bodyStr
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = fmt.Sprintf("%s:%d", IP, halib.DefaultAgentPort)

	m.ServeHTTP(res, req)
	assert.EqualValues(t, http.StatusForbidden, res.Code)
}

func TestRequestStatusManager(t *testing.T) {
	var j []byte
	var err error
	myRSM := &RequestStatusManager{}
	baseTime := time.Date(2017, 9, 11, 15, 4, 5, 0, time.UTC)

	assert.Equal(t, len(myRSM.RequestStatus), 0)
	j, err = json.Marshal(myRSM.GetStatus(baseTime))
	assert.Nil(t, err)
	assert.Equal(t,
		`{"last1":null,"last5":null}`,
		string(j))

	myRSM.Append(baseTime, "/", 200) // [0]
	assert.Equal(t, 1, len(myRSM.RequestStatus))
	assert.Equal(t, baseTime.Unix(), myRSM.RequestStatus[0].When)
	assert.Equal(t, "/", myRSM.RequestStatus[0].URI)
	assert.Equal(t, 1, len(myRSM.RequestStatus[0].Counts))
	assert.Equal(t, uint64(1), myRSM.RequestStatus[0].Counts[200])

	j, err = json.Marshal(myRSM.GetStatus(baseTime))
	assert.Nil(t, err)
	assert.Equal(t,
		`{"last1":[{"url":"/","counts":{"200":1}}],"last5":[{"url":"/","counts":{"200":1}}]}`,
		string(j))

	myRSM.Append(baseTime.Add(30*time.Second), "/", 200)      // [1]
	myRSM.Append(baseTime.Add(30*time.Second), "/", 200)      // [1]
	myRSM.Append(baseTime.Add(30*time.Second), "/", 403)      // [1]
	myRSM.Append(baseTime.Add(30*time.Second), "/proxy", 403) // [2]
	myRSM.Append(baseTime.Add(30*time.Second), "/proxy", 200) // [2]
	assert.Equal(t, 3, len(myRSM.RequestStatus))
	assert.Equal(t, baseTime.Add(30*time.Second).Unix(), myRSM.RequestStatus[1].When)
	assert.Equal(t, baseTime.Add(30*time.Second).Unix(), myRSM.RequestStatus[2].When)
	assert.Equal(t, "/", myRSM.RequestStatus[1].URI)
	assert.Equal(t, "/proxy", myRSM.RequestStatus[2].URI)
	assert.Equal(t, 2, len(myRSM.RequestStatus[1].Counts))
	assert.Equal(t, 2, len(myRSM.RequestStatus[2].Counts))
	assert.Equal(t, uint64(2), myRSM.RequestStatus[1].Counts[200])
	assert.Equal(t, uint64(1), myRSM.RequestStatus[1].Counts[403])
	assert.Equal(t, uint64(1), myRSM.RequestStatus[2].Counts[200])
	assert.Equal(t, uint64(1), myRSM.RequestStatus[2].Counts[403])

	j, err = json.Marshal(myRSM.GetStatus(baseTime.Add(30 * time.Second)))
	assert.Nil(t, err)
	assert.Equal(t,
		`{"last1":[{"url":"/","counts":{"200":3,"403":1}},{"url":"/proxy","counts":{"200":1,"403":1}}],"last5":[{"url":"/","counts":{"200":3,"403":1}},{"url":"/proxy","counts":{"200":1,"403":1}}]}`,
		string(j))

	j, err = json.Marshal(myRSM.GetStatus(baseTime.Add(300 * time.Second)))
	assert.Nil(t, err)
	assert.Equal(t,
		`{"last1":null,"last5":[{"url":"/","counts":{"200":3,"403":1}},{"url":"/proxy","counts":{"200":1,"403":1}}]}`,
		string(j))

	j, err = json.Marshal(myRSM.GetStatus(baseTime.Add(360 * time.Second)))
	assert.Nil(t, err)
	assert.Equal(t,
		`{"last1":null,"last5":null}`,
		string(j))

	myRSM.Append(baseTime.Add(60*time.Second), "/", 200)  // [3]
	myRSM.Append(baseTime.Add(330*time.Second), "/", 200) // [4]
	assert.Equal(t, 5, len(myRSM.RequestStatus))
	assert.Equal(t, baseTime.Add(60*time.Second).Unix(), myRSM.RequestStatus[3].When)
	assert.Equal(t, baseTime.Add(330*time.Second).Unix(), myRSM.RequestStatus[4].When)

	j, err = json.Marshal(myRSM.GetStatus(baseTime.Add(360 * time.Second)))
	assert.Nil(t, err)
	assert.Equal(t,
		`{"last1":[{"url":"/","counts":{"200":1}}],"last5":[{"url":"/","counts":{"200":2}}]}`,
		string(j))

	myRSM.GarbageCollect(baseTime.Add(360*time.Second), 5)

	assert.Equal(t, 2, len(myRSM.RequestStatus))
	assert.Equal(t, baseTime.Add(60*time.Second).Unix(), myRSM.RequestStatus[0].When)
	assert.Equal(t, baseTime.Add(330*time.Second).Unix(), myRSM.RequestStatus[1].When)

	j, err = json.Marshal(myRSM.GetStatus(baseTime.Add(360 * time.Second)))
	assert.Nil(t, err)
	assert.Equal(t,
		`{"last1":[{"url":"/","counts":{"200":1}}],"last5":[{"url":"/","counts":{"200":2}}]}`,
		string(j))
}
