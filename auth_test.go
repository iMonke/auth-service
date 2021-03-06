package main

import (
	"git.gastrodon.io/imonke/monkebase"
	"git.gastrodon.io/imonke/monketype"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

const (
	nick     = "foobar"
	email    = "foo@bar.com"
	password = "foobar2000"
)

var (
	user monketype.User = monketype.NewUser(nick, "", email)
)

func authOK(test *testing.T, auth map[string]interface{}) {
	var key string
	var ok bool
	for _, key = range []string{"token", "secret"} {
		if _, ok = auth[key]; !ok {
			test.Errorf("%s (%v) is not a string", key, auth[key])
		}
	}

	if _, ok = auth["expires"]; !ok {
		test.Errorf("expires (%v) is not an int", auth["expires"])
	}

	var token_bytes []byte
	var err error
	if token_bytes, err = base64.URLEncoding.DecodeString(auth["token"].(string)); err != nil {
		test.Fatal(err)
	}

	if len(token_bytes) != monkebase.TOKEN_LENGTH {
		test.Errorf("token %s is not %d bytes long!", auth["token"], monkebase.TOKEN_LENGTH)
	}

	var secret_bytes []byte
	if secret_bytes, err = base64.URLEncoding.DecodeString(auth["secret"].(string)); err != nil {
		test.Fatal(err)
	}

	if len(secret_bytes) != monkebase.SECRET_LENGTH {
		test.Errorf("secret %s is not %d bytes long!", auth["secret"], monkebase.SECRET_LENGTH)
	}
}

func TestMain(main *testing.M) {
	monkebase.Connect(os.Getenv("MONKEBASE_CONNECTION"))
	monkebase.WriteUser(user.Map())
	monkebase.SetPassword(user.ID, password)
	os.Exit(main.Run())
}

func mustMarshal(it interface{}) (data []byte) {
	var err error
	if data, err = json.Marshal(it); err != nil {
		panic(err)
	}

	return
}

func Test_password(test *testing.T) {
	var data []byte = mustMarshal(map[string]interface{}{
		"email":    email,
		"password": password,
	})

	var request *http.Request
	var err error
	if request, err = http.NewRequest("POST", "https://imonke/auth", bytes.NewReader(data)); err != nil {
		test.Fatal(err)
	}

	var code int
	var r_map map[string]interface{}
	if code, r_map, err = postAuth(request); err != nil {
		test.Fatal(err)
	}

	if code != 200 {
		test.Errorf("password auth did not return 200! status cdoe %d", code)
	}

	authOK(test, r_map["auth"].(map[string]interface{}))
}

func Test_secret(test *testing.T) {
	var secret string
	var err error
	if secret, err = monkebase.CreateSecret(user.ID); err != nil {
		test.Fatal(err)
	}

	var data []byte = mustMarshal(map[string]interface{}{
		"email":  email,
		"secret": secret,
	})

	var request *http.Request
	if request, err = http.NewRequest("POST", "https://imonke/auth", bytes.NewReader(data)); err != nil {
		test.Fatal(err)
	}

	var code int
	var r_map map[string]interface{}
	if code, r_map, err = postAuth(request); err != nil {
		test.Fatal(err)
	}

	if code != 200 {
		test.Errorf("secret auth did not return 200! status cdoe %d", code)
	}

	authOK(test, r_map["auth"].(map[string]interface{}))
}

func Test_badjson(test *testing.T) {
	var request *http.Request
	var err error
	if request, err = http.NewRequest("POST", "https://imonke/auth", bytes.NewReader([]byte("\\`b"))); err != nil {
		test.Fatal(err)
	}

	var code int
	if code, _, err = postAuth(request); err != nil {
		test.Fatal(err)
	}

	if code != 400 {
		test.Errorf("secret auth did not return 400! status cdoe %d", code)
	}
}

func Test_badrequest(test *testing.T) {
	var set []byte
	var sets [][]byte = [][]byte{
		mustMarshal(map[string]interface{}{
			"email": email,
		}),
		mustMarshal(map[string]interface{}{
			"secret": "secret",
		}),
		mustMarshal(map[string]interface{}{
			"password": password,
		}),
		mustMarshal(map[string]interface{}{
			"secret":   "secret",
			"password": password,
		}),
	}

	var request *http.Request
	var code int
	var err error

	for _, set = range sets {
		if request, err = http.NewRequest("POST", "https://imonke/auth", bytes.NewReader(set)); err != nil {
			test.Fatal(err)
		}

		if code, _, err = postAuth(request); err != nil {
			test.Fatal(err)
		}

		if code != 400 {
			test.Errorf("bad request did not return 400! status code %d", code)
		}
	}
}
