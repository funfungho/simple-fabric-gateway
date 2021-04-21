package main

import (
	"io"
	"net/http"
)

const badCRL = `-----BEGIN X509 CRL-----
MIIDFDCCAfwCAQEwDQYJKoZIhvcNAQEFBQAwXzEjMCEGA1UEChMaU2FtcGxlIFNp
Z25lciBPcmdhbml6YXRpb24xGzAZBgNVBAsTElNhbXBsZSBTaWduZXIgVW5pdDEb
MBkGA1UEAxMSU2FtcGxlIFNpZ25lciBDZXJ0Fw0xMzAyMTgxMDMyMDBaFw0xMzAy
MTgxMDQyMDBaMIIBNjA8AgMUeUcXDTEzMDIxODEwMjIxMlowJjAKBgNVHRUEAwoB
AzAYBgNVHRgEERgPMjAxMzAyMTgxMDIyMDBaMDwCAxR5SBcNMTMwMjE4MTAyMjIy
WjAmMAoGA1UdFQQDCgEGMBgGA1UdGAQRGA8yMDEzMDIxODEwMjIwMFowPAIDFHlJ
Fw0xMzAyMTgxMDIyMzJaMCYwCgYDVR0VBAMKAQQwGAYDVR0YBBEYDzIwMTMwMjE4
MTAyMjAwWjA8AgMUeUoXDTEzMDIxODEwMjI0MlowJjAKBgNVHRUEAwoBATAYBgNV
HRgEERgPMjAxMzAyMTgxMDIyMDBaMDwCAxR5SxcNMTMwMjE4MTAyMjUxWjAmMAoG
A1UdFQQDCgEFMBgGA1UdGAQRGA8yMDEzMDIxODEwMjIwMFqgLzAtMB8GA1UdIwQY
MBaAFL4SAcyq6hGA2i6tsurHtfuf+a00MAoGA1UdFAQDAgEDMA0GCSqGSIb3DQEB
BQUAA4IBAQBCIb6B8cN5dmZbziETimiotDy+FsOvS93LeDWSkNjXTG/+bGgnrm3a
QpgB7heT8L2o7s2QtjX2DaTOSYL3nZ/Ibn/R8S0g+EbNQxdk5/la6CERxiRp+E2T
UG8LDb14YVMhRGKvCguSIyUG0MwGW6waqVtd6K71u7vhIU/Tidf6ZSdsTMhpPPFu
PUid4j29U3q10SGFF6cCt1DzjvUcCwHGhHA02Men70EgZFADPLWmLg0HglKUh1iZ
WcBGtev/8VsUijyjsM072C6Ut5TwNyrrthb952+eKlmxLNgT0o5hVYxjXhtwLQsL
7QZhrypAM1DLYqQjkiDI7hlvt7QuDGTJ
-----END X509 CRL-----`

func insertGoodCRL(w http.ResponseWriter, r *http.Request) {
	if err := UpdateCRLOfChannelConfig(badCRL, "mychannel"); err != nil {
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "ok")
}

func insertBadCRL(w http.ResponseWriter, r *http.Request) {
	if err := UpdateCRLOfChannelConfig(badCRL, "mychannel"); err != nil {
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "ok")
}

func cleanCRL(w http.ResponseWriter, r *http.Request) {
	if err := UpdateCRLOfChannelConfig("", "mychannel"); err != nil {
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, "ok")
}

func genValidCRL(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "unimplemented")
}