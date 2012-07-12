package amqp

import (
	"testing"
)

// Test matrix defined on http://www.rabbitmq.com/uri-spec.html
type testURI struct {
	url      string
	username string
	password string
	host     string
	port     int
	vhost    string
}

var uriTests = []testURI{
	testURI{url: "amqp://user:pass@host:10000/vhost",
		username: "user",
		password: "pass",
		host:     "host",
		port:     10000,
		vhost:    "vhost",
	},

	// this fails due to net/url not parsing pct-encoding in host
	// testURI{url: "amqp://user%61:%61pass@ho%61st:10000/v%2fhost",
	//	username: "usera",
	//	password: "apass",
	//	host:     "hoast",
	//	port:     10000,
	//	vhost:    "v/host",
	// },

	testURI{url: "amqp://",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
	},

	testURI{url: "amqp://:@/",
		username: "",
		password: "",
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
	},

	testURI{url: "amqp://user@",
		username: "user",
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
	},

	testURI{url: "amqp://user:pass@",
		username: "user",
		password: "pass",
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
	},

	testURI{url: "amqp://host",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
	},

	testURI{url: "amqp://:10000",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     10000,
		vhost:    defaultURI.Vhost,
	},

	testURI{url: "amqp:///vhost",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    "vhost",
	},

	testURI{url: "amqp://host/",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
	},

	testURI{url: "amqp://host/%2f",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    "/",
	},

	testURI{url: "amqp://host/%2f%2f",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    "//",
	},

	testURI{url: "amqp://[::1]",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "[::1]",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
	},
}

func TestURISpec(t *testing.T) {
	for _, test := range uriTests {
		u, err := ParseURI(test.url)
		if err != nil {
			t.Fatal("Could not parse spec URI: ", test.url, " err: ", err)
		}

		if test.username != u.Username {
			t.Error("For: ", test.url, " usernames do not match. want: ", test.username, " got: ", u.Username)
		}

		if test.password != u.Password {
			t.Error("For: ", test.url, " passwords do not match. want: ", test.password, " got: ", u.Password)
		}

		if test.host != u.Host {
			t.Error("For: ", test.url, " hosts do not match. want: ", test.host, " got: ", u.Host)
		}

		if test.port != u.Port {
			t.Error("For: ", test.url, " ports do not match. want: ", test.port, " got: ", u.Port)
		}

		if test.vhost != u.Vhost {
			t.Error("For: ", test.url, " vhosts do not match. want: ", test.vhost, " got: ", u.Vhost)
		}
	}
}

func TestURIDefaults(t *testing.T) {
	url := "amqp://"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != "amqp://localhost/" {
		t.Fatal("Defaults not encoded properly got:", uri.String())
	}
}

func TestURIComplete(t *testing.T) {
	url := "amqp://bob:dobbs@foo.bar:5678/private"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != url {
		t.Fatal("Defaults not encoded properly want:", url, " got:", uri.String())
	}
}

func TestURIDefaultPortAmqpNotIncluded(t *testing.T) {
	url := "amqp://foo.bar:5672/"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != "amqp://foo.bar/" {
		t.Fatal("Defaults not encoded properly got:", uri.String())
	}
}

func TestURIDefaultPortAmqpsNotIncluded(t *testing.T) {
	url := "amqps://foo.bar:5671/"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != "amqps://foo.bar/" {
		t.Fatal("Defaults not encoded properly got:", uri.String())
	}
}
