package amqp

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"testing"
	"time"
)

func tlsServerConfig() *tls.Config {
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

func tlsClientConfig() *tls.Config {
	cfg := new(tls.Config)
	cfg.RootCAs = x509.NewCertPool()
	cfg.RootCAs.AppendCertsFromPEM([]byte(caCert))

	cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	if err != nil {
		panic(err)
	}

	cfg.Certificates = append(cfg.Certificates, cert)

	return cfg
}

type tlsServer struct {
	net.Listener
	URL      string
	Config   *tls.Config
	Sessions chan *server
}

// Captures the header for each accepted connection
func (s *tlsServer) Serve(t *testing.T) {
	for {
		c, err := s.Accept()
		if err != nil {
			return
		}
		s.Sessions <- newServer(t, c, c)
	}
}

func startTLSServer(t *testing.T, cfg *tls.Config) tlsServer {
	l, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	if err != nil {
		panic(err)
	}

	s := tlsServer{
		Listener: l,
		Config:   cfg,
		URL:      fmt.Sprintf("amqps://%s/", l.Addr().String()),
		Sessions: make(chan *server),
	}
	go s.Serve(t)

	return s
}

// Tests opening a connection of a TLS enabled socket server
func TestTLSHandshake(t *testing.T) {
	srv := startTLSServer(t, tlsServerConfig())
	defer srv.Close()
	go func() {
		select {
		case <-time.After(10 * time.Millisecond):
			t.Fatalf("server timeout waiting for TLS handshake from client")
		case session := <-srv.Sessions:
			session.connectionOpen()
			session.connectionClose()
			session.S.Close()
		}
	}()

	c, err := DialTLS(srv.URL, tlsClientConfig())
	if err != nil {
		t.Fatalf("expected to open a TLS connection, got err: %v", err)
	}
	defer c.Close()

	if st := c.ConnectionState(); !st.HandshakeComplete {
		t.Errorf("expected to complete a TLS handshake, TLS connection state: %+v", st)
	}
}

const caCert = `
-----BEGIN CERTIFICATE-----
MIIC0TCCAbmgAwIBAgIUW418AvO6YD2WD5X/coo9geXvauEwDQYJKoZIhvcNAQEL
BQAwEzERMA8GA1UEAwwITXlUZXN0Q0EwHhcNMjIwNDA0MDMxNjIxWhcNMzIwNDAx
MDMxNjIxWjATMREwDwYDVQQDDAhNeVRlc3RDQTCCASIwDQYJKoZIhvcNAQEBBQAD
ggEPADCCAQoCggEBAOU1K7mS1z7U9mdDaxpmaj/JwDjtIquqGv1IE4hi9RU/jkse
2GYUuOf5Viu7LMaKCS4I5qwyXT/yPUviw5b1bjEviNURCmQScvXKJg9/P31JlZPc
RBV6Hnrku7nZeJcDfiAo8YDA3QPdr1uhTjjnIo4x2SYJfKDgBvsLfxXETUkFsECQ
uuBLl3VL8jeU8y9wAg0y4gmcOZUyKPlcr9dsmDKftzkas+Zd4JR6e327U+VvVxv4
hDMQGxx/yirKxMwH2+pwkYOWwOTb1FGi8iVjdAwgjlKsLqUR9eSN2AXW/v41E79h
Sb9uEGeIfscGOrgVfVtVRZWNPDzM1/DpYSpsP2ECAwEAAaMdMBswDAYDVR0TBAUw
AwEB/zALBgNVHQ8EBAMCAQYwDQYJKoZIhvcNAQELBQADggEBAJWlbsQqPQPX/lOF
CUnta6OHpt6OlPnN1lUhvbNQVpt0KdmuHQKwSyGiq6fVDt//2RcxPwCYO9AKUZZn
/T48+4haGFEKLRVQrDX7/CQoGbHwWbbgcNpJiQ3XPhGvUrLy6VoahAYK81D23XT0
b4TobCYY+8ny5qhyQgBZ7Jme2jDk0MXt8yjZFGcyA0fRy54ql1AxIXdx5/yvq2CI
nLGEZQLsYhz/r+4gooIkLOIFD2JKoW9p56T52XiRxmz0tAs4WFW38Z6VrzB48Lpq
dBH0sqzECTVQmX8er1taHa+Tg29tuXyqNkBn0tB9qq7G24SDaQFqpW7+J1kNnAY2
KEcj4Ic=
-----END CERTIFICATE-----
`

const serverCert = `
-----BEGIN CERTIFICATE-----
MIIC8zCCAdugAwIBAgIBATANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDDAhNeVRl
c3RDQTAeFw0yMjA0MDQwMzE2MjFaFw0zMjA0MDEwMzE2MjFaMCUxEjAQBgNVBAMM
CTEyNy4wLjAuMTEPMA0GA1UECgwGc2VydmVyMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEArH6lT2bGHiGLaWh+eXLXeAcbXq2bGDyeS1zsRGT/+D6YBzFo
7FAUMbN8iJtjWNBlRRdVyiUJn5TUYHrXdCeoAs5JrdzNefax+3toDu6I+nJG4dO3
nbZ9zo8q7uBZtCi/Oph28CtwQxMjNuf3pHlSSKZG5/br4w6HhgSnuLBxyxpr3CrU
2kdHZGIGjX4sEGJ/aDCKIUem/jEXp5i1n13+p5QmPfZc6KX2+sUmN6todqlZpsEV
WZAYLXshHJvViLMTrCJQ3KfqoF7o4PrwQJoQ3jlNct2ZF68H0Iz/bT/VQfPS+4CL
t71xHhU/mYn+kIt1w0APhCMMkYPOhijmGlDAbwIDAQABo0AwPjAJBgNVHRMEAjAA
MAsGA1UdDwQEAwIFIDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHREECDAGhwR/
AAABMA0GCSqGSIb3DQEBCwUAA4IBAQA8dM07fsOfxU7oVWQN3Aey0g8/wc2eL0ZY
oxxvKJqZjOSamNEiwb+lTTmUGg78wiIpFK4ChpPbMpURcRVrxTJuWRg4qagFWn18
RGNqNVnucFSOPe3e+zZqnjWc22PDbBt4B3agdpJsHq4hpdOdJreAh/7MMrWyFMZI
PZ5WFjThFpeQw+vArYUq6BMD3H86G++gfWlSv0yPxWG+Q44+tfZSyiiM6k0+p3nw
1k0eBIA+BibcZ+dQu0m2Vq7XKxfOeN67LY8aElGNSRmH/YAyo5mSrrvzbbZKpr4/
cn4f79jDcBEtbMeCav0eHXzSwWWpteZvdLOm6rxnHd3jngRPqRj+
-----END CERTIFICATE-----
`

const serverKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEArH6lT2bGHiGLaWh+eXLXeAcbXq2bGDyeS1zsRGT/+D6YBzFo
7FAUMbN8iJtjWNBlRRdVyiUJn5TUYHrXdCeoAs5JrdzNefax+3toDu6I+nJG4dO3
nbZ9zo8q7uBZtCi/Oph28CtwQxMjNuf3pHlSSKZG5/br4w6HhgSnuLBxyxpr3CrU
2kdHZGIGjX4sEGJ/aDCKIUem/jEXp5i1n13+p5QmPfZc6KX2+sUmN6todqlZpsEV
WZAYLXshHJvViLMTrCJQ3KfqoF7o4PrwQJoQ3jlNct2ZF68H0Iz/bT/VQfPS+4CL
t71xHhU/mYn+kIt1w0APhCMMkYPOhijmGlDAbwIDAQABAoIBAQCFfEw5MfNHBfZ4
z+Bv46tSu00262oGS4LEF1jPZMmhNe84QchMd3vpKljI7lbnN/3mhbRiBl94Gxhu
wSFSRg4CfdkOrrxkEcCSOGHCjF18UksAH3MMnVimLKywxvUkMhQqKCqCmVr6zSiH
KOO/aBOBHQvqHm9U+r1tvNR+XCzzWmz0OavKquXhdTbLyHWeR7rS4mn53Up0hQIe
la8PzWpw+mcNz00UbeeK84tUro4+wGV6aFHh23F+RqpXqW8VNnBS0zVtqYc7x7a8
WGcxlLt1hWeiFgEJUuui2i1YxyHDb3IyhGckVG+lSzTpgjl9Pisl7jjlVBRAkiWw
fRJfWuyBAoGBAOERpoRSpXCBum3T7ZwcD3pWHgNY1za9zqsgZayumlhqvxmia77o
HvXyroj+8DTUZ+NiKn0HnF8Hni/2fhAWs9XPlELM9BENQP5kcJzcBSalUOa23C6V
Iba30BB8I1v8AhYLdCFdpio+aQnd1S4VH4sTRoQjsspOSesxDWa1snJ/AoGBAMQz
VadEoBR8ZHsZjks7JaU2G1h2tdU3hfynD8DPq0SFnbS3NhVphmYAdfbOrea9ilmQ
AyO/TJD6xmK7IELWYNtS4DCqAaQKxfOsS8wF97lJ1/XKDNQND8PfbVwzLF8NcbMT
G15Vq9dkJpiGBe0YYxdjtPtwM+FSriK37YIyidoRAoGAQqGBFKeLBvXBBYa6T38X
LfaUyBTjEfe7WXor36WJWCeyD5rAHzKFB/ciqLgg0OMZJn4HaiB4sMGGmVh2FblC
4Eel8ujOUMYFucpudGHGvJwwiT0VjkzkQD3GwTqfFTpUO8aESOR6rwLvAdbEp/Hk
9r1sIO6Ynb/zrkdFWmTsQW0CgYBvXrhjH2hC2K1s1v/XonZnBoSVPaVPp5nN5cLi
br9IQRRZLZpsox7gLajIdV9vV+39kurFUuSSc1dDWfchGXGXbb7GwOn3hQoCnK3V
3RlWOx10bsHDaLqnM99u87lfJ1GAFft2G+lUdYwXDhS1Fh/Beh6Uj4dTgsxH9uHC
AxAPEQKBgCkOBQBfPg4mZdYKAaNoRvNhJr1XdruQnFsBCdvCm2x8DAkNGme7Ax8H
QYB+jH1eC/uInPUnW3kh7KPqgdGVNbYeZGT1FwuwoJEubR6+rBiRm3HsjBJlR+Tv
RK8SwjKS0ZkCb8eujkyBakXdF2tAY9Du8HODID+CuuxhGD1YsF2q
-----END RSA PRIVATE KEY-----
`

const clientCert = `
-----BEGIN CERTIFICATE-----
MIIC4jCCAcqgAwIBAgIBAjANBgkqhkiG9w0BAQsFADATMREwDwYDVQQDDAhNeVRl
c3RDQTAeFw0yMjA0MDQwMzE2MjFaFw0zMjA0MDEwMzE2MjFaMCUxEjAQBgNVBAMM
CTEyNy4wLjAuMTEPMA0GA1UECgwGY2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAtL3BQX/ubTjO+m4WMDe7OiPquHkyE/G8zHNoFXH/pjcf0iYq
lBWD8+bmoPxzOy2RtRawZIR5ZtHkPH8YZues3nwnnUPYQfUGlqlSc4TOyO7vn0Eq
ufJZCVF3+DWTHaV41K20/cpLuCs308KKVt7XosDmj4iwc1t7NcjS8TFI4/SpBIYg
TzII+RzFXhT3FFXuo5ZkOoti0IyUXxALX/Ba+ubR9vrJC4BS7VtoSQQ/uRr5x9MB
/yZWUDTcOOJNjh1mKe+9vEFAjCgRFOwaZzph38Av/L2e4uPTsjTX6MsMQZK1nl5h
03HoQZ3FRJLVKTHVszihIODgNQ5ReM3YBxrqiwIDAQABoy8wLTAJBgNVHRMEAjAA
MAsGA1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjANBgkqhkiG9w0BAQsF
AAOCAQEAxMJjHzZ6BiSwzieH+Ujz70fU+qa4xdoY8tt32gnBJt9d+dsgjaAIeGZR
veAA3Ak0aM+1zoCGc2792oEv7gOz946brbkJLNC5e43ZlDFHYbj1OYcMtxL4CCyW
2QImY/CgoXfVkth1SSRl+wGEwMlS8XuriczwU4Sl/faJHE+mwwJNlNBNKjv1GoUb
0T3lnfuUNoXzH1SstR/6+/eD0CNvTsTvkC9eXJULt2V93jTFNGlM6KcVvF5E7BDm
iSYtaQ8vtPF64LKM+dF3ymzJDoT6EyYU6t1X5RPQ6fQTRbCvZ74NsWoqWzA4VB20
rgNKRkWBHGxKqNMrF1YQvm2hWxw6pA==
-----END CERTIFICATE-----
`

const clientKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAtL3BQX/ubTjO+m4WMDe7OiPquHkyE/G8zHNoFXH/pjcf0iYq
lBWD8+bmoPxzOy2RtRawZIR5ZtHkPH8YZues3nwnnUPYQfUGlqlSc4TOyO7vn0Eq
ufJZCVF3+DWTHaV41K20/cpLuCs308KKVt7XosDmj4iwc1t7NcjS8TFI4/SpBIYg
TzII+RzFXhT3FFXuo5ZkOoti0IyUXxALX/Ba+ubR9vrJC4BS7VtoSQQ/uRr5x9MB
/yZWUDTcOOJNjh1mKe+9vEFAjCgRFOwaZzph38Av/L2e4uPTsjTX6MsMQZK1nl5h
03HoQZ3FRJLVKTHVszihIODgNQ5ReM3YBxrqiwIDAQABAoIBAQCmfc2R2pj1P8lZ
4yLJU+1CB2fmeq3otVvnMcAFUTfgExNa8BF0y8T7Xg3A6gvzzWxVVgsy7N0wG9SU
7ba6xFr3r4KGWcLSLzXcfykWhJY/fep51vvWwinGbaeHm0JjotQFheYdisXpZtZM
WP46O5iDshIw0gdInFKJHu9BgtbUNDRAAI9xsmDcITSXf7ZblH5CCLJ6QLFn1Olg
O11k6ABYmIxKSXfLylFDzUxfnGhXkrv8/sQwj3B7XxBYVoAF24uw2ufZ+qIlcxUy
m/T/IU1lOLNmfWwDk2+t7zslOpXAEZtrni7u2GF54wIbMr/Z0unGC1B0AUXDwp8y
MNODQdcBAoGBAOe+p5neqgoUL8kBlRvp/dp5kIsVPNBN7+vbVYVzeIEBNixfYe9S
0dixQHiJr29sWn/4yC2KkPQewyDVBf7iNXlcXfkjbEjU0cZaMJao4sKki/N/GF64
IpC8HRM1BSvaSnrJZtZ6aWljpvWRa8bGRdQWb397ufChNpstYEALMMZLAoGBAMeo
hGZJcWIEtl1oJIHsKFKf7jwftFidzyy2ShVLX52n3HkbCYMdwg5DIjQ6Vj9GMy8E
AvrcQduhLw4otdPL+X270xX8WKbxRAvlDnGfmOX1+z/lJy2OLi/HpZeUyM6Oe/Uy
ijaGTKOqRSi9xpJOdDR17e/fcYKzvPejNxZcDMTBAoGAeH/VLBfweI8oja8J9lrE
CX7eXsNrPLDZyNziaiKxjPqxTX9HMCbzQGZiLIsDMr+3iwU0KSH830LDmWXK2U6M
GY+iuXHm0zP949Jvo1crmaPvtWvnoxDBwFpgD+Woy7WUtqXUmD9MYmVToiq8TL45
/t6vmS0fcPSSrTt56bMn6GMCgYEAvRDLL8FkaRllR9aSm6VyGavxAWZUdYYa5ZBJ
XxjdFoIauWPtAghv9umDvklv2sMzPNZjrAJfKwfbc2EBrep9+56dKTipCo11jn39
y4MCWuEwZzUsgGsfOYepO31dGpy6rVqKn09Vy7Y1f3sWSv2X9QWnp3rEFqz1yNr6
E2ZfgQECgYA943tAbNFmAZW3WdS82cpM+HXOkKuYkPfFsBq/eoKoi1p0WLtueQZ3
cC39vheaWdRJPI/dKjhgifY6pBcVLzGIf9gD6VKVbIrcx6U3uMULEQb3GqvdIOtU
d5WpU0bwFMa+vYfrlAjngXlDW/tGqK8ietb6n+xW15M1w4mrEFcAng==
-----END RSA PRIVATE KEY-----
`
