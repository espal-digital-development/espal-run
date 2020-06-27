// +build windows

package openssl

// Stub; Windows 10 has openssl by default

func (o *OpenSSL) checkAndInstall() error {
	return nil
}
