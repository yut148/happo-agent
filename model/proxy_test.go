package model

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostToAgent1(t *testing.T) {
	const STUB_RESPONSE = "OK"

	ts := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, STUB_RESPONSE)
			}))
	defer ts.Close()
	re, _ := regexp.Compile("([a-z]+)://([A-Za-z0-9.]+):([0-9]+)(.*)")
	found := re.FindStringSubmatch(ts.URL)
	host := found[2]
	port, _ := strconv.Atoi(found[3])

	json_data := []byte("{}")
	status_code, response, err := postToAgent(host, port, "test", json_data)
	assert.EqualValues(t, status_code, http.StatusOK)
	assert.Contains(t, response, STUB_RESPONSE)
	assert.Nil(t, err)
}
