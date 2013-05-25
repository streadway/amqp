package amqp_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"net"
	"testing"
	"time"
)

type tlsServer struct {
	net.Listener
	URL    string
	Config *tls.Config
	Header chan []byte
}

// Captures the header for each accepted connection
func (s *tlsServer) Serve() {
	for {
		c, err := s.Accept()
		if err != nil {
			return
		}

		header := make([]byte, 4)
		io.ReadFull(c, header)
		s.Header <- header
		c.Write([]byte{'A', 'M', 'Q', 'P', 0, 0, 0, 0})
		c.Close()
	}
}

func tlsConfig() *tls.Config {
	cfg := new(tls.Config)

	cfg.ClientCAs = x509.NewCertPool()
	cfg.ClientCAs.AppendCertsFromPEM([]byte(caCert))

	cert, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		panic(err)
	}

	cfg.Certificates = append(cfg.Certificates, cert)
	cfg.ClientAuth = tls.RequireAndVerifyClientCert

	return cfg
}

func startTlsServer() tlsServer {
	cfg := tlsConfig()

	l, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	if err != nil {
		panic(err)
	}

	s := tlsServer{
		Listener: l,
		Config:   cfg,
		URL:      fmt.Sprintf("amqps://%s/", l.Addr().String()),
		Header:   make(chan []byte, 1),
	}

	go s.Serve()
	return s
}

// Tests that the server has handshaked the connection and seen the client
// protocol announcement.  Does not nest that the connection.open is successful.
func TestTLSHandshake(t *testing.T) {
	srv := startTlsServer()
	defer srv.Close()

	cfg := new(tls.Config)
	cfg.RootCAs = x509.NewCertPool()
	cfg.RootCAs.AppendCertsFromPEM([]byte(caCert))

	cert, _ := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	cfg.Certificates = append(cfg.Certificates, cert)

	_, err := amqp.DialTLS(srv.URL, cfg)

	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatalf("did not succeed to handshake the TLS connection after 10ms")
	case header := <-srv.Header:
		if string(header) != "AMQP" {
			t.Fatalf("expected to handshake a TLS connection, got err: %v", err)
		}
	}
}

const caCert = `
-----BEGIN CERTIFICATE-----
MIICxjCCAa6gAwIBAgIJAJ9B7ynYqS1oMA0GCSqGSIb3DQEBBQUAMBMxETAPBgNV
BAMTCE15VGVzdENBMB4XDTEzMDExNjIxMjgxMFoXDTE0MDExNjIxMjgxMFowEzER
MA8GA1UEAxMITXlUZXN0Q0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
AQDFZnUFUtu/Q1QW5SrDO5Saq1rpgkeBJdGN5emWGvH8lMVr/NN22FHUuMbmOubW
UG809Z537WJp8v4m/h5Vbv9H3fOld+h8B0THqdh9W9Kv2s0MhQGLW4o6GCQmACUx
yMBUfiN6GxYhq8pgJTApgAKQbSxldedZDR9GJSIDnFpp7GQwlrslKtskwwqkQ51O
S/tHxusIdVvGx0EgJHIQb2qegsVaFcGJsQG719inIMy+EtCDUcxJoWROIcuE8t4y
1RVZtZySDAcIvZWW7XQdro6tqKY3P4FBsDgqCdOF5L5uMNZVa9EUY6LEtHBBjcX9
oA7DLULUNNIXbZrpiGBfuB75AgMBAAGjHTAbMAwGA1UdEwQFMAMBAf8wCwYDVR0P
BAQDAgEGMA0GCSqGSIb3DQEBBQUAA4IBAQB3BsY5m1XOB34OAcq/+q8uvCWbluIh
vmQek8aOE0d+/uP3dGOCfZOoaAopn2e4bIEHeYqMpCdw1tH/S73VJvZEbruIfmr6
pnZHbXz6WBL+PKTRU6fkP7U8iEM13tFsd37QaNUzlkCeh00SqHKkRdT6EIJIbwqV
sXytD6aTNCehAOt1VijGOQcJVuIOFJP5aa2pzANgO2OUVRnivB3MfVSE0oVnzKdd
tnbX68Hah+Naa0WVoRp1FN5wG9nOR9v7kZMYgoQZYmkdq2b21Jmj0ZAEkYiwIUDa
Jk3Pu0K26zPToz+0UP0djv2VfvcufhY+Y+5vWmB2/2eI9WsZkPxaMyGF
-----END CERTIFICATE-----
`

// Assumes subject CN is 127.0.0.1 and SAN includes 127.0.0.1
const serverCert = `
-----BEGIN CERTIFICATE-----
MIIC8zCCAdugAwIBAgIBBjANBgkqhkiG9w0BAQUFADATMREwDwYDVQQDEwhNeVRl
c3RDQTAeFw0xMzA0MTUxMzEyNDBaFw0xNDA0MTUxMzEyNDBaMCUxEjAQBgNVBAMM
CTEyNy4wLjAuMTEPMA0GA1UECgwGc2VydmVyMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAynXINa1PsGDP5uxEnK6Av2CGIBlIsCptltg/NJxPxfwOB/9y
q2z6tfpgIL4oc/OB5gEv+M+6egNdmL+ASFKvMktwyNncbScT2rXGMYQmQhd5Jnnj
VINq2Yt+tN4z07xYmbFo87TaFgQprfmjK7W6A5xq95gvNMLbL4pv/SyBRXPFb2CS
2aiVhN4rVsrbzJKUBXoTyUVSESxL5kK93zx+faddY+7XTZpUUtekeKYlkInS0ouU
pizjeZRPThOZePP9EiMmTl0FHgW9cb9GtlKemNEQvqZfEi2y64/K861VL1pwMxNM
S9GUXqzy4A6kgRqKy9Yn1+j/yp2iNTtj2VUKKQIDAQABo0AwPjAJBgNVHRMEAjAA
MAsGA1UdDwQEAwIFIDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHREECDAGhwR/
AAABMA0GCSqGSIb3DQEBBQUAA4IBAQAoowlI2X5w7kjQACkXcQEtNqiQjUiBarH2
2AwpNFuPANavhN2Y3hWbo84l3RhU3Yds286p3TY0MFPDCT9jUtdWp8PmYKoG8PQi
JFnEK6x15MOtyJmxgQu5nG6b3mqbFSk7ZKKj1oqAeYfuUeOdMqfbEdYNXjUwvcIL
j0jWDhh9dXmeCh7E6iF5SqnaEFCxv3vOvKgdapLx2Yf6n1dNN57U0hmhFse8Pnv1
CXlKvEIdKs2PXxboP0ulHNxVt5wU+XbzxTg7k6nwzXXoUjZJJFX21vWGSn+PZdDS
HclQTpS6rEzjmKUvxFRmVbU3FI/DqqBhrQDyThfsqeXxIsslhd9S
-----END CERTIFICATE-----
`

const serverKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAynXINa1PsGDP5uxEnK6Av2CGIBlIsCptltg/NJxPxfwOB/9y
q2z6tfpgIL4oc/OB5gEv+M+6egNdmL+ASFKvMktwyNncbScT2rXGMYQmQhd5Jnnj
VINq2Yt+tN4z07xYmbFo87TaFgQprfmjK7W6A5xq95gvNMLbL4pv/SyBRXPFb2CS
2aiVhN4rVsrbzJKUBXoTyUVSESxL5kK93zx+faddY+7XTZpUUtekeKYlkInS0ouU
pizjeZRPThOZePP9EiMmTl0FHgW9cb9GtlKemNEQvqZfEi2y64/K861VL1pwMxNM
S9GUXqzy4A6kgRqKy9Yn1+j/yp2iNTtj2VUKKQIDAQABAoIBAEAx0GWUqmvWhpVF
3QuBGTmVNXIAElgpW840ivX3iiPQo/JNQOKyD1ycIta+9LyvPUTco4VU+F+vqYHB
Vr+X2A0udnh4+7dwaI80i78vk6HpJ3TXuQkXEk4gOPDIc85zLfStmAWOFohckYqk
WOSHHo/+jLws+OrVzgHo91FjRynp1v1vQqx7/PEiQaMZFfP8++qe6hGqbPKcGEFE
8Dqxm7oSvFNpGh1Mb/EkF5VMkB/oj52Zr5yXNvMaKCvERkBTXiljIE8GWOW8uWxx
RwxT59ZkCltg9apDRqpD1w/ERj/dsUHDRVMgnvuYx84ctzrCmE0WbJ1i5qMY+gk3
+7+srbECgYEA6+u2/DIZS0C4nLhzy/80iS/bHXly4voO6H3dNK9PCJ3Uf1Ldzohh
Q62RHHxGmDQDQ92v7w2tKEGXl19ewvZK2BXJujl/RaXpFOFdlMG3pcF8oq00Xbyn
zryMzbgxhk8W7XMSKNcZ13fJqzrHk68MI6Pcw1tbhVcqt/qI1h4DYGMCgYEA27EF
GMCCzh7m/S6XzN6HfIYmUhwfhaWdR3NFu6cPPEWmu+P1nefiH4IXKnzwKoDvCLI3
yDXPLGHPihR/TAgJZ0Fr4W+BWjOYZ/7H2uLHkRUTq1P6sKBEMS+4UQKTVpYZzr5f
LCYLuyyrKM+mgD6xLiUuWHcKGjxDGedLQTT7QwMCgYBfFsfPSKYXRcPjLxlFPNzA
+q/3Zk8fGyjNHoX9STeywmK22wCZ0TBa5edaMuEFUdmLDhxuXvXPBvkBwyffrwOl
qsp/K9OXj/KtPtTIM3hA8Aa7TtKPgY2lbyvVcwtLFi+ojzvfiCtbRGXdhTiR1vku
mEpP10/BI8wNEYb7vmcf/wKBgHTZoOZbbreHhEDiCWe0bMf06mj+AF4rio44Z7y5
zMa4HUkOpNOKRKGRphS5Q1y4G2u6ryTUSg9HLwY4hMTB+Y5sI59SmbCKhOO7hj2M
Ja2rUjzsfAh6Fgs7YIPmJMwHJk7qvuBSlCbISXl5iQvpTIBI0m/HUR5HM0GR3lse
fQ5VAoGAU2x5qdsq2ITGr+XycZYYVA07WsN+LkomtND0iDBFBWRDCCSRJHLPaN9d
HcZWqEaVVO27mIq6P/PBS0VT0rOkOuti/BVWC4gjfsBG050whnES78xoAce5x0XG
QKWExCP1p7f4ZzYYY3/NELpYLDhol+PwhhNgSyCesFpJuvZCY04=
-----END RSA PRIVATE KEY-----
`

const clientCert = `
-----BEGIN CERTIFICATE-----
MIIC4jCCAcqgAwIBAgIBBDANBgkqhkiG9w0BAQUFADATMREwDwYDVQQDEwhNeVRl
c3RDQTAeFw0xMzAxMjMyMjA5MTlaFw0xNDAxMjMyMjA5MTlaMCUxEjAQBgNVBAMM
CTEyNy4wLjAuMTEPMA0GA1UECgwGY2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAuKjq7EQKNEsps9r0x/ed09eqasUSOPsNx7SVoHkDq2BkE8xW
uLch2ifTA37wKvzFd1DeF9iARpt3n2a0etisMDOnZvWOLh14qVT8YJ6VPnppfkMA
q7sRq539jDHd5E7VYniPOmWldRduC96pVzgsbbdpMUW8az4kVJXcDp7qSMuXG13Y
A/SYVWcGt9hfPyDJOR1kz8btJkAHGYVS2gK9TCewOYwBo8kBPQxT4A+D8eR9zZEh
lraF8k0y5RGVUwocUG/U3gPgyrG+OIk9xeO8pNRh/QhgT9y5XZl1eCiJSRSXT3C+
FyQxTs7k3vMIWAylTH1AmCGEty08yTD+JF7hnwIDAQABoy8wLTAJBgNVHRMEAjAA
MAsGA1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjANBgkqhkiG9w0BAQUF
AAOCAQEAi6bJ9lyUqunl2KXKHn5Lne59YBqJ3XEqFlN/9H5MrMheOusoDtduzTYj
rm2hCY84ZceKZdz2F5vsz1XakX2apbPVBOVWPOUylyVWlRRGB7PSXsSHyU8CXI96
2HtpK/WiWuVOXwdxIX/MlyL5dx8ckvje0YXMJYS1c+OGEmeRBNMBitVRI3ZhcjAw
SIj71EiXG1EsF2Ok7H1bXiEgjmFr4Ui4CpmeITybGAqoVeA9+t09FFFwABS1wlOM
ZGcOWPhEGv6vCpaCLLdAaU9K+BYnBhtVIhNpt5Q1gtj/pLauVWa2Ouhcq41l8HPR
tKIEURQFA5C0JL68IPorTQt5z+D+bg==
-----END CERTIFICATE-----
`

const clientKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAuKjq7EQKNEsps9r0x/ed09eqasUSOPsNx7SVoHkDq2BkE8xW
uLch2ifTA37wKvzFd1DeF9iARpt3n2a0etisMDOnZvWOLh14qVT8YJ6VPnppfkMA
q7sRq539jDHd5E7VYniPOmWldRduC96pVzgsbbdpMUW8az4kVJXcDp7qSMuXG13Y
A/SYVWcGt9hfPyDJOR1kz8btJkAHGYVS2gK9TCewOYwBo8kBPQxT4A+D8eR9zZEh
lraF8k0y5RGVUwocUG/U3gPgyrG+OIk9xeO8pNRh/QhgT9y5XZl1eCiJSRSXT3C+
FyQxTs7k3vMIWAylTH1AmCGEty08yTD+JF7hnwIDAQABAoIBAGtcUlGRcXlb5eAa
wkxsy8c50WwILgMQ+78LYB8PnLGL9kOIfzcfyj/C/a0/pTTpB4nKa4XjqxjiFNeJ
aA8wYFQaBA8ZX1OycM/KiH1IVi8gDquJGx+9QJXN4ncbGw49Q1TgES37oQoF2EZw
a3Y5Q6N6il9KUzTqyUagZnPdswskzMMFBPn0aYn//O6CMd3ptOZx31/+fy1eaWeO
dh0joT1cmoAQXCHTcvbsiijO8ytmVsZlHWT9B3nKX5QFxhVub3rhVSxg7I1N8hzE
hR0wNz8PkU2Eyo31DOBg3TULzd9xVSrdgCFx9VJFL8de72J+81X3kttrpbtK4ojb
xzwfNBECgYEA82zWJxQkpsq+LGpVt4tU504XZp6xuCC67VIv3sxdT5hWbbqQ6uw8
Uw5zY4byQpVU6dvZ+k7ztwqzOeO5Yu2xknvBXWCmNhlEz6j/tD9VNvSYUuar1Q3m
5/b5xK+ju5zVuD+/4hLbOPcG0KZ9D3a+fQJz4RkT16MeeUKeFbs4A8cCgYEAwjLx
Ojm28K+Dsv0dSCfskeQ0nV7Us0yUNKj8lNzxz2J4I1ZVg3M3lyC9fo8cdnFwXmKT
ZD6WUjAd3Rfo+B36IPiS3cM9LsJ5W8mTRD67evogmqqUoXADK0QzA0S9TQlIqAHr
fjK5PyfeGjJGX1DwmEhZ0P6CWc7bvKd8gCbWA2kCgYEA281V3jmREs6VQ/PMbIyy
YJ4iATagkPt07qA8u3hbdWi/+hrxij9ABVtSE/ehP0AqSXSMcjniVVCjH02ic1Lf
+b4njxKbYtQUT1JxeieJ4bKg7JJ/bEU+UAyx4ckbFmh6jwF5WUDflKNyEuuSl2kI
fka9re7//MG83Y+qwUKpRLcCgYEAkWRmajtPlb8yEM2kIKOTYF7EbZXUFTEePJbQ
E/ufJq8IVxyKBVI7qnAeryQiISMpB+ExjHm3PW08zozaJPj8jbbM7i8AHYQILAos
sYlt/9JImsNfZ8Ze+QOkVawfNg/fT7mwP9lmC7yjcmV1fmMw3jI83FXP7cELjqCu
e5uX2xECgYEAi1W7xOXM8wsKBkDA3XfPfnOCsqL0Jlo0gcAUe8xk8J1zCXRlJbJK
LGfvLUFr9oyYJuScT0lloTdkA72CvTMHozSsBMGdWPw5/e6nXdbUJeaqcsvZLotW
fuqiAigY1BT+sTwiBOKsh/Df8Gf5SRS80l23AmznWgdxWyZ/mYIB1zk=
-----END RSA PRIVATE KEY-----
`
