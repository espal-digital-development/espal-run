package cockroach

const (
	minAllowedPortRange = 1024
	maxAllowedPortRange = 49151

	defaultPortStart     = 26257
	defaultHTTPPortStart = 26080

	secondsIntervalBetweenNodesStart = 10

	cockroachCreatingNewDatabase          = "Creating a new cockroach database.."
	cockroachResettingDatabase            = "Resetting the cockroach database.."
	cockroachResettingDatabaseNotRequired = "No database found. Skipping reset.."

	httpUserSQL = `
CREATE USER %s WITH PASSWORD '%s';
GRANT admin to %s;
`

	setupDatabaseSQL = `CREATE DATABASE app;

CREATE USER selecter;
GRANT SELECT ON DATABASE app TO selecter;

CREATE USER creator;
GRANT SELECT ON DATABASE app TO creator;
GRANT CREATE ON DATABASE app TO creator;

CREATE USER inserter;
GRANT SELECT ON DATABASE app TO inserter;
GRANT INSERT ON DATABASE app TO inserter;

CREATE USER updater;
GRANT SELECT ON DATABASE app TO updater;
GRANT UPDATE ON DATABASE app TO updater;

CREATE USER deletor;
GRANT SELECT ON DATABASE app TO deletor;
GRANT DELETE ON DATABASE app TO deletor;

CREATE USER migrator;
GRANT GRANT ON DATABASE app TO migrator;
GRANT CREATE ON DATABASE app TO migrator;
GRANT DROP ON DATABASE app TO migrator;
GRANT SELECT ON DATABASE app TO migrator;
GRANT INSERT ON DATABASE app TO migrator;
GRANT UPDATE ON DATABASE app TO migrator;
GRANT DELETE ON DATABASE app TO migrator;
`
)
