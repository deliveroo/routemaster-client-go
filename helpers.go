package routemaster

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
)

// Panics if err is not nil.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Reads JSON formatted bytes from an io.Reader to create a map.
func readJSON(r io.Reader) map[string]interface{} {
	var v map[string]interface{}
	bytes, _ := ioutil.ReadAll(r)
	json.Unmarshal(bytes, &v)
	return v
}

// Pretty prints json.
func prettyJSON(v interface{}) string {
	buf, _ := json.MarshalIndent(v, "", "  ")
	return string(buf)
}

// validates that s is a valid https absolute url.
func isValidAbsoluteURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}
