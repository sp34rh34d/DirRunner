# DirRunner

DirRunner is a Go-based scanner for DNS enumeration, directory discovery, virtual hosts, FUZZ payload testing, and basic HTTP fingerprinting.

It runs as a command-line tool with worker pools, streaming result output, reusable HTTP transports, JSON output, deterministic exports, optional Tor routing, and no third-party Go dependencies.

## Build

```bash
go build -o dirrunner ./cmd/dirrunner && chmod +x dirrunner
```

## General Usage

```bash
./dirrunner <module> [flags]
```

Every module can read its main target from stdin by using `-` as a positional argument:

```bash
echo https://example.com | ./dirrunner dir -
echo example.com | ./dirrunner dns -
echo https://example.com/FUZZ | ./dirrunner fuzz - -w wordlist/payloads.txt
echo https://192.0.2.10 | ./dirrunner vhost - -d example.com
echo https://example.com | ./dirrunner fingerprint -
```

Modules:

- `dns`: subdomain DNS enumeration.
- `dir`: directory enumeration. File checks are enabled only when `--ext` / `-e` is used.
- `vhost`: virtual host enumeration using the `Host` header.
- `fuzz`: FUZZ placeholder enumeration in URL, headers, cookies, user-agent, basic auth, JSON body, or POST form data.
- `fingerprint`: HTTP headers and basic technology fingerprinting.

## Global Flags

These flags are available in all modules.

| Flag | Description |
| --- | --- |
| `--threads / -t N` | Number of concurrent workers. Default is `10`. |
| `--json / -j` | Print results as JSON. |
| `--export / -o FILE` | Write final results to a file. |
| `--verbose / -v` | Print detailed diagnostic output. |

`--wordlist / -w FILE` is available in `dns`, `dir`, `vhost`, and `fuzz`.

## Shared HTTP Request Flags

These flags are available in `dir`, `vhost`, `fuzz`, and `fingerprint`.

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. |
| `--timeout / -T DURATION` | HTTP request timeout, for example `5s` or `1500ms`. |
| `--user-agent / -A VALUE` | Custom `User-Agent` header. |
| `--cookie / -C VALUE` | Cookie header value. |
| `--header / -H 'Name: value'` | Custom HTTP header. Can be repeated. |
| `--username / --user / -U VALUE` | HTTP Basic Auth username. |
| `--password / --pass / -P VALUE` | HTTP Basic Auth password. |
| `--insecure / -k` | Disable TLS certificate validation. |
| `--follow-redirects / -L` | Follow HTTP redirects. |
| `--tor / -Q` | Route HTTP traffic through a SOCKS5 Tor proxy. |
| `--tor-proxy / -q HOST:PORT` | Tor SOCKS5 proxy address. Default is `127.0.0.1:9050`. |

Response filtering flags are available in `dir`, `vhost`, and `fuzz`:

| Flag | Description |
| --- | --- |
| `--exclude-size / -S LIST` | Hide exact response sizes, for example `3244,777`. |
| `--exclude-size-range / -R LIST` | Hide inclusive size ranges, for example `1000-2000,9000-9500`. |

`--method / -X` and `--codes / -c` are used by `dir` and `fuzz`.

## DNS Module

Use `dns` to resolve subdomains from a wordlist.

### DNS Flags

| Flag | Description |
| --- | --- |
| `--domain / -d DOMAIN` | Base domain to enumerate. |
| `--wordlist / -w FILE` | Subdomain wordlist. |

### DNS Examples

Basic DNS enumeration:

```bash
./dirrunner dns -d example.com -w wordlist/subdomains.txt
```

Increase concurrency and export results:

```bash
./dirrunner dns \
  --domain example.com \
  --wordlist wordlist/subdomains.txt \
  --threads 30 \
  --export dns-results.txt
```

Verbose DNS output prints only the detailed discovery line:

```bash
./dirrunner dns -d example.com -w wordlist/subdomains.txt --verbose
```

## Dir Module

Use `dir` to enumerate directories and, optionally, files. Directory checks always run. File checks run only when `--ext` / `-e` is provided.

Normal `dir` output prints only the discovered path, not the full URL:

```text
directory  200     777      /cgi-bin/
file       200     1820     /index.php
```

### Dir Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Base target URL. |
| `--wordlist / -w FILE` | Main directory wordlist. Alias for the directory wordlist. |
| `--dir-wordlist FILE` | Directory wordlist path. |
| `--file-wordlist / -f FILE` | File-name wordlist path. |
| `--ext / -e LIST` | File extensions to test. Enables file enumeration. Example: `php,txt,bak`. |
| `--method / -X METHOD` | HTTP method. Default is `GET`. |
| `--codes / -c LIST` | Status codes to show. Default is `200,301,302`. |
| `--recursive / -r` | Recursively scan discovered directories. |
| `--depth / -D N` | Maximum recursion depth. Default is `1`. |
| `--wildcard-check / -W` | Check missing-path wildcard responses before enumeration. Disabled by default. |
| `--skip-wildcard-check` | Legacy compatibility flag to skip wildcard checks. |

### Dir Examples

Directory enumeration:

```bash
./dirrunner dir \
  -u https://example.com \
  -w wordlist/directory.txt \
  -c 200,301,302,403
```

Read the target URL from stdin:

```bash
echo https://example.com | ./dirrunner dir -
```

Directory and file enumeration at the same time:

```bash
./dirrunner dir \
  -u https://example.com \
  -w wordlist/directory.txt \
  -f wordlist/filename.txt \
  -e php,txt,bak \
  -c 200,301,302,403
```

Recursive directory scan:

```bash
./dirrunner dir \
  --url https://example.com \
  --wordlist wordlist/directory.txt \
  --recursive \
  --depth 3 \
  --codes 200,301,302,403
```

Exclude noisy response sizes:

```bash
./dirrunner dir \
  -u https://example.com \
  -w wordlist/directory.txt \
  --exclude-size 777,1338728 \
  --exclude-size-range 9000-9500
```

Route the scan through Tor:

```bash
./dirrunner dir \
  -u https://example.com \
  -w wordlist/directory.txt \
  --tor
```

Tor Browser usually exposes SOCKS5 on `127.0.0.1:9150`:

```bash
./dirrunner dir \
  -u https://example.com \
  -w wordlist/directory.txt \
  --tor \
  --tor-proxy 127.0.0.1:9150
```

## VHost Module

Use `vhost` to test virtual hosts by sending each wordlist entry as a `Host` header value.

### VHost Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL or IP address that receives the request. |
| `--domain / -d DOMAIN` | Domain appended to each wordlist entry to build the `Host` header. |
| `--wordlist / -w FILE` | Hostname wordlist. |

### VHost Examples

Enumerate vhosts against an IP:

```bash
./dirrunner vhost \
  -u https://192.0.2.10 \
  -d example.com \
  -w wordlist/subdomains.txt
```

Use HTTP filters:

```bash
./dirrunner vhost \
  -u https://192.0.2.10 \
  -d example.com \
  -w wordlist/subdomains.txt \
  --exclude-size 3244 \
  --timeout 3s
```

## Fuzz Module

Use `fuzz` wherever a value should be replaced by words from the wordlist. Put `FUZZ` in the URL, headers, cookies, user-agent, Basic Auth credentials, JSON body, or POST form data.

When fuzzing headers, Basic Auth, cookies, user-agent, JSON, or POST data, the normal output prints the FUZZ value instead of repeating the unchanged URL.

### Fuzz Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. It can contain `FUZZ`, or be paired with another FUZZ option. |
| `--wordlist / -w FILE` | Wordlist used to replace `FUZZ`. |
| `--method / -X METHOD` | HTTP method. Default is `GET`; body flags force `POST`. |
| `--codes / -c LIST` | Status codes to show. Default is `200,301,302`. |
| `--json-body / -b JSON` | JSON request body containing `FUZZ`. |
| `--data / -p FORM` | POST form data containing `FUZZ`, for example `username=FUZZ&password=fake`. |
| `--base64 / -B` | Base64 encode FUZZ values before sending. |
| `--urlencode / -E` | URL encode FUZZ values before sending. |
| `--username / --user / -U VALUE` | Basic Auth username. Can contain `FUZZ`. |
| `--password / --pass / -P VALUE` | Basic Auth password. Can contain `FUZZ`. |
| `--header / -H 'Name: value'` | Custom header. Header name or value can contain `FUZZ`. |
| `--cookie / -C VALUE` | Cookie header. Can contain `FUZZ`. |
| `--user-agent / -A VALUE` | User-Agent header. Can contain `FUZZ`. |

If both `--base64` and `--urlencode` are enabled, base64 is applied first and URL encoding second.

### Fuzz Examples

Fuzz a URL path:

```bash
./dirrunner fuzz \
  -u 'https://example.com/FUZZ' \
  -w wordlist/paths.txt \
  -c 200,301,302,403
```

Fuzz a query parameter:

```bash
./dirrunner fuzz \
  -u 'https://example.com/search?q=FUZZ' \
  -w wordlist/payloads.txt \
  -c 200
```

Fuzz POST form data:

```bash
./dirrunner fuzz \
  -X POST \
  -u https://example.com/login \
  -p 'username=FUZZ&password=fake' \
  -w username.txt \
  --exclude-size 3244
```

Fuzz a POST JSON field:

```bash
./dirrunner fuzz \
  -u https://example.com/api/login \
  -b '{"username":"admin","password":"FUZZ"}' \
  -w wordlist/passwords.txt \
  -c 200,401,403
```

Fuzz Basic Auth username:

```bash
./dirrunner fuzz \
  -u https://example.com/admin \
  --user FUZZ \
  --pass fake \
  -w username.txt
```

Fuzz Basic Auth password:

```bash
./dirrunner fuzz \
  -u https://example.com/admin \
  -U admin \
  -P FUZZ \
  -w wordlist/passwords.txt
```

Fuzz a custom header:

```bash
./dirrunner fuzz \
  -u https://example.com/ \
  -H 'X-Forwarded-Host: FUZZ.example.com' \
  -w wordlist/subdomains.txt \
  -c 200,302
```

Fuzz a cookie value:

```bash
./dirrunner fuzz \
  -u https://example.com/profile \
  -C 'session=FUZZ' \
  -w wordlist/tokens.txt \
  -c 200,302
```

Encode FUZZ values:

```bash
./dirrunner fuzz \
  -u 'https://example.com/search?q=FUZZ' \
  -w wordlist/payloads.txt \
  --base64 \
  --urlencode
```

## Fingerprint Module

Use `fingerprint` to inspect response status, size, server headers, security headers, and basic technology hints from the target.

### Fingerprint Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. |
| `--timeout / -T DURATION` | HTTP request timeout. |
| `--user-agent / -A VALUE` | Custom User-Agent header. |
| `--cookie / -C VALUE` | Cookie header value. |
| `--header / -H 'Name: value'` | Custom HTTP header. Can be repeated. |
| `--username / --user / -U VALUE` | HTTP Basic Auth username. |
| `--password / --pass / -P VALUE` | HTTP Basic Auth password. |
| `--insecure / -k` | Disable TLS certificate validation. |
| `--follow-redirects / -L` | Follow HTTP redirects. |
| `--tor / -Q` | Route HTTP traffic through Tor. |
| `--tor-proxy / -q HOST:PORT` | Tor SOCKS5 proxy address. |

### Fingerprint Examples

Basic fingerprint:

```bash
./dirrunner fingerprint -u https://example.com
```

Fingerprint with custom headers and redirects:

```bash
./dirrunner fingerprint \
  -u https://example.com \
  -H 'X-Forwarded-For: 127.0.0.1' \
  --follow-redirects \
  --verbose
```

## Output Control

Print JSON:

```bash
./dirrunner dir -u https://example.com -w wordlist/directory.txt --json
```

Export results:

```bash
./dirrunner dns -d example.com -w wordlist/subdomains.txt --export dns-results.txt
```

Verbose mode:

```bash
./dirrunner fuzz \
  -u https://example.com/login \
  -p 'username=FUZZ&password=fake' \
  -w username.txt \
  --verbose
```

## Notes

- `dir` starts dispatching requests while the wordlist is being read.
- `dir` always enumerates directories; file enumeration only starts when `--ext` / `-e` is present.
- `--exclude-size` and `--exclude-size-range` are useful when a target returns same-sized false positives.
- `--tor` requires a running SOCKS5 Tor proxy.
- For module-specific help, run `./dirrunner <module> -h`.
