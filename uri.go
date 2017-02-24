// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var errURIScheme = errors.New("AMQP scheme must be either 'amqp://' or 'amqps://'")

var schemePorts = map[string]int{
	"amqp":  5672,
	"amqps": 5671,
}

var defaultURI = URI{
	Scheme:   "amqp",
	Host:     "localhost",
	Port:     5672,
	Username: "guest",
	Password: "guest",
	Vhost:    "/",
}

// URI represents a parsed AMQP URI string.
type URI struct {
	Scheme   string
	Host     string
	Port     int
	Username string
	Password string
	Vhost    string
}

// ParseURI attempts to parse the given AMQP URI according to the spec.
// See http://www.rabbitmq.com/uri-spec.html.
//
// Default values for the fields are:
//
//   Scheme: amqp
//   Host: localhost
//   Port: 5672
//   Username: guest
//   Password: guest
//   Vhost: /
//
func ParseURI(uri string) (URI, error) {

	me := defaultURI
	paten := `^amqp(s)?://[^:\/\/]+:[^:\/\/]+@[^:\/\/]+:[0-9]+\/[^:\/\/]*`
	reg, _ := regexp.Compile(paten)
	if !reg.Match([]byte(uri)) {
		return defaultURI, errors.New("unvalid")
	}
	piceInfo := strings.Split(uri, ":")
	scheme := piceInfo[0]
	userName := piceInfo[1][2:]
	passwordHostNameInfo := strings.Split(piceInfo[2], "@")
	password := passwordHostNameInfo[0]
	hostName := passwordHostNameInfo[1]
	portVhostInfo := strings.Split(piceInfo[3], "/")
	port, _ := strconv.Atoi(portVhostInfo[0])
	var vhost string
	if portVhostInfo[1] == "" {
		vhost = "/"
	} else {
		vhost = portVhostInfo[1]
	}

	me.Scheme = scheme
	me.Host = hostName
	me.Port = port
	me.Username = userName
	me.Password = password
	me.Vhost = vhost
	return me, nil

}

// Splits host:port, host, [ho:st]:port, or [ho:st].  Unlike net.SplitHostPort
// which splits :port, host:port or [host]:port
//
// Handles hosts that have colons that are in brackets like [::1]:http
func splitHostPort(addr string) (host, port string) {
	i := strings.LastIndex(addr, ":")

	if i >= 0 {
		host, port = addr[:i], addr[i+1:]

		if len(port) > 0 && port[len(port)-1] == ']' && addr[0] == '[' {
			// we've split on an inner colon, the port was missing outside of the
			// brackets so use the full addr.  We could assert that host should not
			// contain any colons here
			host, port = addr, ""
		}
	} else {
		host = addr
	}

	return
}

// PlainAuth returns a PlainAuth structure based on the parsed URI's
// Username and Password fields.
func (uri URI) PlainAuth() *PlainAuth {
	return &PlainAuth{
		Username: uri.Username,
		Password: uri.Password,
	}
}

func (uri URI) String() string {
	var authority string

	if uri.Username != defaultURI.Username || uri.Password != defaultURI.Password {
		authority += uri.Username

		if uri.Password != defaultURI.Password {
			authority += ":" + uri.Password
		}

		authority += "@"
	}

	authority += uri.Host

	if defaultPort, found := schemePorts[uri.Scheme]; !found || defaultPort != uri.Port {
		authority += ":" + strconv.FormatInt(int64(uri.Port), 10)
	}

	var vhost string
	if uri.Vhost != defaultURI.Vhost {
		vhost = uri.Vhost
	}

	return fmt.Sprintf("%s://%s/%s", uri.Scheme, authority, url.QueryEscape(vhost))
}
