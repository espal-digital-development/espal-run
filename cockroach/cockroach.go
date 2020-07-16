package cockroach

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/juju/errors"
)

const cockroachNotFoundInstalling = "Did not find `cockroach`. Attempting to installing.."

type Cockroach struct {
	desiredNodes  int
	host          string
	rootUser      string
	portStart     int
	httpHost      string
	httpUser      string
	httpPassword  string
	httpPortStart int
	databasePath  string
	caKeyName     string

	resetDB bool

	certsDir string
	safeDir  string

	rePortListen *regexp.Regexp
}

// GetDesiredNodes gets desiredNodes.
func (c *Cockroach) GetDesiredNodes() int {
	return c.desiredNodes
}

// SetDesiredNodes sets desiredNodes.
func (c *Cockroach) SetDesiredNodes(desiredNodes int) {
	c.desiredNodes = desiredNodes
}

// GetHost gets host.
func (c *Cockroach) GetHost() string {
	return c.host
}

// SetHost sets host.
func (c *Cockroach) SetHost(host string) {
	c.host = strings.TrimSpace(host)
}

// GetRootUser gets rootUser.
func (c *Cockroach) GetRootUser() string {
	return c.rootUser
}

// SetRootUser sets rootUser.
func (c *Cockroach) SetRootUser(rootUser string) {
	c.rootUser = strings.TrimSpace(rootUser)
}

// GetPortStart gets portStart.
func (c *Cockroach) GetPortStart() int {
	return c.portStart
}

// SetPortStart sets portStart.
func (c *Cockroach) SetPortStart(portStart int) {
	c.portStart = portStart
}

// GetHTTPHost gets httpHost.
func (c *Cockroach) GetHTTPHost() string {
	return c.httpHost
}

// SetHTTPHost sets httpHost.
func (c *Cockroach) SetHTTPHost(httpHost string) {
	c.httpHost = strings.TrimSpace(httpHost)
}

// GetHTTPUser gets httpUser.
func (c *Cockroach) GetHTTPUser() string {
	return c.httpUser
}

// SetHTTPUser sets httpUser.
func (c *Cockroach) SetHTTPUser(httpUser string) {
	c.httpUser = strings.TrimSpace(httpUser)
}

// GetHTTPPassword gets httpPassword.
func (c *Cockroach) GetHTTPPassword() string {
	return c.httpPassword
}

// SetHTTPPassword sets httpPassword.
func (c *Cockroach) SetHTTPPassword(httpPassword string) {
	c.httpPassword = strings.TrimSpace(httpPassword)
}

// GetHTTPPortStart gets httpPortStart.
func (c *Cockroach) GetHTTPPortStart() int {
	return c.httpPortStart
}

// SetHTTPPortStart sets httpPortStart.
func (c *Cockroach) SetHTTPPortStart(httpPortStart int) {
	c.httpPortStart = httpPortStart
}

// GetDatabasePath gets databasePath.
func (c *Cockroach) GetDatabasePath() string {
	return c.databasePath
}

// SetDatabasePath sets databasePath.
func (c *Cockroach) SetDatabasePath(databasePath string) error {
	var err error
	c.databasePath, err = filepath.Abs(filepath.FromSlash(
		strings.TrimRight(strings.TrimSpace(databasePath), "/")))
	if err != nil {
		return errors.Trace(err)
	}
	c.certsDir = filepath.FromSlash(c.databasePath + "/certs")
	c.safeDir = filepath.FromSlash(c.databasePath + "/safe")
	return nil
}

// GetResetDB gets resetDB.
func (c *Cockroach) GetResetDB() bool {
	return c.resetDB
}

// SetResetDB sets resetDB.
func (c *Cockroach) SetResetDB(resetDB bool) {
	c.resetDB = resetDB
}

func (c *Cockroach) isSafePortRange(port int, fieldName string) error {
	if port < minAllowedPortRange {
		return errors.Errorf("%s cannot be in the range of reserved OS ports", fieldName)
	}
	if port > maxAllowedPortRange {
		return errors.Errorf("%s cannot be over the range of application-safe ports", fieldName)
	}
	return nil
}

// New returns a new instance of c.
func New() (*Cockroach, error) {
	c := &Cockroach{
		desiredNodes:  1,
		host:          "localhost",
		rootUser:      "root",
		portStart:     defaultPortStart,
		httpHost:      "localhost",
		httpPortStart: defaultHTTPPortStart,

		caKeyName: "ca.key",
	}

	var err error
	c.rePortListen, err = regexp.Compile(`(?m)^\s*cockroach\s+(\d+).*?\(LISTEN\)\s*$`)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return c, nil
}
