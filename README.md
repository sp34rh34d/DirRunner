# DirRunner

DirRunner is a Go-based scanner for DNS enumeration, directory discovery, virtual hosts, FUZZ payload testing, basic HTTP fingerprinting, path-traversal detection, 401/403 bypass probing, SQL-injection indicators, web crawling, and site cloning.

The web-assessment modules (`traversal`, `bypass`, `sqli`, `crawl`, `clone`) are a Go port of the WebRunner toolkit, integrated into DirRunner's architecture: the same worker pools, streaming output, reusable HTTP transport, JSON output, exports, and an optional HTTP/SOCKS5 proxy.

## Build

```bash
go build -o dirrunner ./cmd/dirrunner && chmod +x dirrunner
```

Cross-compiles cleanly (pure Go, standard library only):

```bash
GOOS=windows GOARCH=amd64 go build -o dirrunner.exe ./cmd/dirrunner
GOOS=linux   GOARCH=amd64 go build -o dirrunner      ./cmd/dirrunner
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
echo 'https://example.com/?file=FUZZ' | ./dirrunner traversal -
echo https://example.com/admin | ./dirrunner bypass -
echo 'https://example.com/item?id=1' | ./dirrunner sqli -
echo https://example.com | ./dirrunner crawl -
echo https://example.com | ./dirrunner clone -
```

Every URL module (`dir`, `vhost`, `fuzz`, `fingerprint`, `traversal`, `bypass`, `sqli`, `crawl`, `clone`) accepts multiple targets from a file with `--url-file FILE` (one URL per line), and each target is scanned in turn with the results aggregated. Piping several URLs with `-` works the same way.

Modules:

- `dns`: subdomain DNS enumeration.
- `dir`: directory enumeration. File checks are enabled only when `--ext` / `-e` is used.
- `vhost`: virtual host enumeration using the `Host` header.
- `fuzz`: FUZZ placeholder enumeration in URL, headers, cookies, user-agent, basic auth, JSON body, or POST form data.
- `fingerprint`: HTTP headers and basic technology fingerprinting.
- `traversal`: path-traversal / local file disclosure scanner. Payloads are generated from the `wordlist/traversal-*.txt` lists.
- `bypass`: 401/403 access-control bypass probe using header tricks and path mutations.
- `sqli`: SQL-injection indicator scanner for query parameters and request headers.
- `crawl`: web crawler that discovers URLs, JS/image assets, e-mails, and regex matches (unifies WebRunner scraping, email-extractor, and regx).
- `clone`: download a site breadth-first, following same-domain links and assets, saving every resource to disk.

## Global Flags

These flags are available in all modules.

| Flag | Description |
| --- | --- |
| `--threads / -t N` | Number of concurrent workers. Default is `40`. |
| `--json / -j` | Print results as JSON. |
| `--export / -o FILE` | Write final results to a file. |
| `--verbose / -v` | Print detailed diagnostic output. |

`--wordlist / -w FILE` is available in `dns`, `dir`, `vhost`, and `fuzz`. In `sqli` it overrides the payload wordlist, and in `bypass` the header wordlist.

## Shared HTTP Request Flags

These flags are available in `dir`, `vhost`, `fuzz`, `fingerprint`, `traversal`, `bypass`, `sqli`, `crawl`, and `clone`.

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. |
| `--timeout / -T DURATION` | HTTP request timeout, for example `5s` or `1500ms`. |
| `--user-agent / -A VALUE` | Custom `User-Agent` header. |
| `--random-agent / -a` | Use a random `User-Agent` per request. |
| `--cookie / -C VALUE` | Cookie header value. |
| `--header / -H 'Name: value'` | Custom HTTP header. Can be repeated. |
| `--username / --user / -U VALUE` | HTTP Basic Auth username. |
| `--password / --pass / -P VALUE` | HTTP Basic Auth password. |
| `--insecure / -k` | Disable TLS certificate validation. |
| `--follow-redirects / -L` | Follow HTTP redirects. |
| `--proxy / -x URL` | Route traffic through a proxy: `http://`, `https://`, `socks5://` or `socks5h://`. For Tor use `socks5h://127.0.0.1:9050`. |

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
| `--codes / -c LIST` | Status codes to show. Default is `200,301,302,401,403`. |
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

Route the scan through a proxy (Burp, ZAP, or Tor):

```bash
./dirrunner dir \
  -u https://example.com \
  -w wordlist/directory.txt \
  --proxy http://127.0.0.1:8080
```

For Tor, point `--proxy` at its SOCKS5 port (Tor Browser usually exposes `127.0.0.1:9150`):

```bash
./dirrunner dir \
  -u https://example.com \
  -w wordlist/directory.txt \
  --proxy socks5h://127.0.0.1:9150
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
| `--csrf-field NAME` | Refresh this anti-CSRF token (and its session cookie) before each request — required for CSRF-protected forms. |
| `--csrf-url URL` | Page to GET the fresh token from. Default: the target URL. |
| `--encode LIST` | Chain encodings over the FUZZ value in order (`base64`, `urlencode`, `xor`); only the encoded result is sent. Same engine as `traversal`/`sqli`. |
| `--xor-key KEY` | XOR key used with `--encode xor`. Default is `DirRunner`. |
| `--base64 / -B` | Shortcut for `--encode base64`. |
| `--urlencode / -E` | Shortcut: append `urlencode` to the chain. |
| `--username / --user / -U VALUE` | Basic Auth username. Can contain `FUZZ`. |
| `--password / --pass / -P VALUE` | Basic Auth password. Can contain `FUZZ`. |
| `--header / -H 'Name: value'` | Custom header. Header name or value can contain `FUZZ`. |
| `--cookie / -C VALUE` | Cookie header. Can contain `FUZZ`. |
| `--user-agent / -A VALUE` | User-Agent header. Can contain `FUZZ`. |

`--encode` chains encodings as a pipeline in the given order and sends **only** the encoded FUZZ value (not the plain word): `--encode xor,base64,urlencode` sends `urlencode(base64(xor(word)))`, for targets that decode the parameter before using it. The `-B`/`-E` shortcuts still work (base64 first, then urlencode) and are equivalent to `--encode base64,urlencode`.

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

Fuzz a CSRF-protected login (fresh token + session cookie fetched before each request):

```bash
./dirrunner fuzz \
  -u https://example.com/login \
  -p 'username=admin&password=FUZZ&csrf=x' \
  --csrf-field csrf \
  -w wordlist/passwords.txt \
  -c 302
```

The value of the `csrf` field in `--data` is a placeholder; DirRunner overwrites it with a fresh token (read from the login page, or from `--csrf-url`) and injects the matching session cookie on every request. Point `--csrf-field` at the real field name and keep the fuzzed field (here `password`) separate. The same works with a JSON body (`-b '{"username":"admin","password":"FUZZ","csrf":"x"}'`).

Encode FUZZ values:

```bash
./dirrunner fuzz \
  -u 'https://example.com/search?q=FUZZ' \
  -w wordlist/payloads.txt \
  --encode base64,urlencode

# XOR then base64, with a custom key
./dirrunner fuzz \
  -u 'https://example.com/search?q=FUZZ' \
  -w wordlist/payloads.txt \
  --encode xor,base64 --xor-key s3cr3t
```

### Fuzz Scenarios

Where to put the `FUZZ` marker for each injection point. Add `-c LIST` to pick which status codes count as a hit, `-v` to see every request, and `-j` for JSON output.

| Scenario | Command |
| --- | --- |
| **URL path** | `fuzz -u 'https://t/FUZZ' -w paths.txt -c 200,301,403` |
| **Query parameter** | `fuzz -u 'https://t/search?q=FUZZ' -w payloads.txt -c 200` |
| **POST form field** | `fuzz -u https://t/login -p 'username=FUZZ&password=fake' -w users.txt` |
| **POST JSON field** | `fuzz -u https://t/api/login -b '{"username":"admin","password":"FUZZ"}' -w pass.txt` |
| **CSRF-protected login** | `fuzz -u https://t/login -p 'username=admin&password=FUZZ&csrf=x' --csrf-field csrf -w pass.txt -c 302` |
| **CSRF, token from another page** | `fuzz -u https://t/login -p 'username=admin&password=FUZZ&csrf=x' --csrf-field csrf --csrf-url https://t/login-form -w pass.txt` |
| **Basic Auth username / password** | `fuzz -u https://t/admin --user FUZZ --pass fake -w users.txt` |
| **Custom header** | `fuzz -u https://t/ -H 'X-Forwarded-Host: FUZZ.t' -w subdomains.txt -c 200,302` |
| **Cookie value** | `fuzz -u https://t/profile -C 'session=FUZZ' -w tokens.txt -c 200,302` |
| **Encoded payload** | `fuzz -u 'https://t/search?q=FUZZ' -w payloads.txt --encode base64,urlencode` |

Notes:
- For a CSRF form, the `csrf` value in `--data`/`--json-body` is a placeholder: DirRunner overwrites it with a fresh token (from the target, or `--csrf-url`) and injects the matching session cookie on every request, while fuzzing the other field (here `password`).
- The token field name after `--csrf-field` must match the real hidden field (`csrf`, `_token`, `authenticity_token`, …).

## Fingerprint Module

Use `fingerprint` to inspect response status, size, server headers, security headers, and basic technology hints from the target.

### Fingerprint Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. |
| `--timeout / -T DURATION` | HTTP request timeout. |
| `--user-agent / -A VALUE` | Custom User-Agent header. |
| `--random-agent / -a` | Use a random User-Agent per request. |
| `--cookie / -C VALUE` | Cookie header value. |
| `--header / -H 'Name: value'` | Custom HTTP header. Can be repeated. |
| `--username / --user / -U VALUE` | HTTP Basic Auth username. |
| `--password / --pass / -P VALUE` | HTTP Basic Auth password. |
| `--insecure / -k` | Disable TLS certificate validation. |
| `--follow-redirects / -L` | Follow HTTP redirects. |
| `--proxy / -x URL` | Route traffic through an `http`/`https`/`socks5`/`socks5h` proxy. |

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

## Traversal Module

Use `traversal` to probe for path-traversal / local file disclosure. Payloads are generated by combining traversal sequences with file-disclosure targets at each depth, loaded from `wordlist/traversal-*.txt`.

Detection is **signature-based**: a hit is confirmed when the response body contains a leaked-file fingerprint (for example `root:...:0:0:`, `[boot loader]`, a private key), loaded from `wordlist/traversal-signatures.txt`. This avoids the false positives that response-length comparison produces on dynamic pages. Pass `--regx PATTERN` to additionally report (and show) any response matching a custom word or regex — useful for a file whose contents you know but that has no built-in signature.

Because detection is signature-based, the default file list is a single small **canary** set (`wordlist/traversal-files.txt`, 10 Linux + Windows files such as `/etc/passwd`, `win.ini`, `boot.ini`) that are almost always present and covered by a signature — enough to confirm a vulnerability without multiplying thousands of requests. Every run tests the whole list; there is no OS switch. The exhaustive WebRunner list ships as `wordlist/traversal-files-full.txt` for exploitation once a hole is confirmed; point `--files-wordlist` (or `--custom-file`) at it when needed.

Put `FUZZ` in the URL to inject the payload at a specific spot (for example a query parameter); otherwise the payload is appended to the URL path directory. No path prefix is prepended by default; `--custom-path` adds prefixed payload variants (e.g. `cgi-bin/../../etc/passwd`).

`--encode` chains encodings as a pipeline in the given order and sends **only** the encoded result (not the plain payload): `--encode xor,base64,urlencode` sends `urlencode(base64(xor(payload)))`, intended for targets that decode the parameter.

### Traversal Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. May contain a `FUZZ` marker. |
| `--url-file FILE` | File with target URLs, one per line. |
| `--min-depth N` / `--max-depth N` | Traversal depth range. Defaults are `1` and `3`. |
| `--strings-wordlist FILE` | Traversal sequence wordlist. Default `wordlist/traversal-strings.txt`. |
| `--files-wordlist FILE` | File-disclosure list (all payloads, both OSes). Default `wordlist/traversal-files.txt` (full list: `traversal-files-full.txt`). |
| `--signature-wordlist FILE` | Leaked-file content signatures (one regex per line). Default `wordlist/traversal-signatures.txt`. |
| `--regx PATTERN` | Also report (and show) responses matching this regex/word, on top of signature detection. |
| `--custom-path LIST` | Comma-separated prefix paths prepended to every payload (e.g. `cgi-bin/`). Default: no prefix. |
| `--custom-traversal-string LIST` | Comma-separated traversal strings (overrides the strings wordlist). |
| `--custom-file LIST` | Comma-separated file-disclosure targets (overrides the files wordlist). |
| `--encode LIST` | Chain encodings in order (`base64`, `urlencode`, `xor`); only the encoded payload is sent, and `200`/`401`/`403` responses are reported. |
| `--xor-key KEY` | XOR key used with `--encode xor`. Default is `DirRunner`. |
| `--continue-on-success` | Keep scanning a target after the first finding. |

### Traversal Examples

```bash
./dirrunner traversal -u 'https://example.com/download.php?file=FUZZ'
./dirrunner traversal -u https://example.com/page --max-depth 6 --encode base64,urlencode
./dirrunner traversal --url-file targets.txt --continue-on-success
```

## Bypass Module

Use `bypass` to probe a `401`/`403`-protected resource with header tricks and path mutations. Base candidates are loaded from `wordlist/bypass-headers.txt` and `wordlist/bypass-paths.txt`. A candidate is reported only when it reaches a **`2xx` success** that was not already reachable at baseline — a change from `403` to `404` or `500` is not treated as a bypass, which keeps the output free of that noise.

**URL-override headers** (`X-Original-URL`, `X-Rewrite-URL`, `X-Override-URL`, `X-HTTP-DestinationURL`, `X-Original-Url`) are front-end routing headers whose *value* is the path the request is routed to. The access-control check often runs on the real request line while the back-end serves the header's path — so the bypass is to request an **allowed** path (not the protected one) while pointing the override header at the protected path. For a target of `.../admin`, `bypass` automatically requests a random dummy path (which normally `404`s) with `X-Original-URL: /admin` and friends; a `2xx` means the override served the protected content. This is the "URL-based access control can be circumvented" technique — sending the header *to* `/admin` keeps the `403`, so it must be sent to a different path.

### Bypass Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. |
| `--url-file FILE` | File with target URLs, one per line. |
| `--header-wordlist / -w FILE` | Base bypass header wordlist (`Name: value` per line). |
| `--path-wordlist FILE` | Base path-suffix wordlist. |
| `--bypass-header 'Name: value'` | Extra bypass header to test. Can be repeated. |
| `--bypass-path SRC` | Extra path, or a file/directory of paths, to test. |
| `--continue-on-success` | Keep scanning a target after the first finding. |

### Bypass Examples

```bash
./dirrunner bypass -u https://example.com/admin
./dirrunner bypass -u https://example.com/admin --bypass-header 'X-Real-IP: 127.0.0.1'
./dirrunner bypass --url-file protected.txt --continue-on-success
# URL-override: auto-tries X-Original-URL/X-Rewrite-URL: /admin against a dummy path
./dirrunner bypass -u https://example.com/admin
```

## SQLi Module

Use `sqli` to test query parameters, request headers **and POST bodies** (form or JSON — e.g. login forms) for SQL-injection indicators using three detectors:

- **Error-based**: database error strings (`wordlist/sqli-errors.txt`) or a new `5xx` where the baseline was below `500`.
- **Boolean-based**: a TRUE condition that leaves the page like the baseline while the paired FALSE condition changes it (length or status differential). This replaces the old raw length-delta check, which fired on ordinary dynamic content.
- **Time-based**: `SLEEP`/`WAITFOR`/`pg_sleep` payloads whose induced delay is measured against the baseline latency and confirmed with a second request. Controlled by `--sleep N` (seconds; `0` disables it).
- **Auth/login bypass**: a payload that makes the response newly redirect (`3xx`) where the baseline did not — the classic login-bypass signal (e.g. `username=administrator'--` comments out the password check and the app redirects to the account page). Redirect following stays off by default so the `3xx` is observed; the payload list includes `administrator'--`, `admin'--`, `' OR 1=1-- -`, etc.

Payloads, header names, and error patterns are loaded from `wordlist/sqli-*.txt`. Boolean and time-based detection run on query parameters and body fields; header testing is error-based.

**NoSQL operator injection** (MongoDB auth-bypass labs): for a **JSON** body (`-b`), sqli also replaces every string field with Mongo operator objects at once (`{"$ne":null}`, `{"$gt":""}`, `{"$regex":".*"}`), the classic operator-injection probe — a status/length change (often a `500`) flags it. Note these labs use a **JSON** login, so use `-b`, not `--data` (a form body returns `400`):

```bash
./dirrunner sqli -u https://TARGET/login -b '{"username":"admin","password":"admin"}' --continue-on-success
# -> body [NoSQL operator, all fields] | {"$ne":null} | NoSQL operator: server error changed
```

To test a **login form** (or any POST endpoint), pass the body with `--data` (form-urlencoded) or `--json-body` (JSON); POST is implied. If the body contains a `FUZZ` marker, only that field is injected; otherwise every field (all string fields for JSON) is tested in turn:

```bash
./dirrunner sqli -u https://TARGET/login -p 'username=admin&password=admin'        # tests both fields
./dirrunner sqli -u https://TARGET/login -p 'username=FUZZ&password=admin'         # tests only username
./dirrunner sqli -u https://TARGET/api/login -b '{"user":"admin","pass":"admin"}'  # tests both JSON fields
```

**CSRF-protected forms**: when the form carries an anti-CSRF token (a per-session, often single-use hidden field), a static token is rejected before the request reaches the database. Pass `--csrf-field NAME` and DirRunner fetches a fresh token (and its session cookie) with a `GET` before every request, injects both, and keeps the CSRF field out of injection. The field's value in `--data` is then irrelevant:

```bash
./dirrunner sqli -u https://TARGET/login -X POST \
  --data 'csrf=x&username=FUZZ&password=admin' --csrf-field csrf --continue-on-success
```

**Blind injection in a cookie or header** (e.g. a `TrackingId` cookie, conditional-response labs): put a `FUZZ` marker in the **cookie** (`-C`), a header (`-H`), or the URL, and sqli injects there with all detectors — including boolean, which catches the small TRUE/FALSE response difference. Use the channel the app actually reads (a `TrackingId` cookie is `-C`, **not** `-H`). For a cookie whose value starts with `FUZZ`, sqli automatically fetches a server-issued value and uses it as the injection base, so you don't need the real token:

```bash
./dirrunner sqli -u https://TARGET/ -C 'TrackingId=FUZZ' --continue-on-success           # auto-fetches the token
./dirrunner sqli -u https://TARGET/ -C 'TrackingId=<realtoken>FUZZ' --continue-on-success # or supply it
./dirrunner sqli -u https://TARGET/ -H 'X-Forwarded-For: <ip>FUZZ' --continue-on-success
```

**DBMS fingerprinting**: `--fingerprint` identifies the backend engine (and version where it can be leaked) instead of running the full injection scan. It tries several techniques per engine until one works: **UNION** (version reflected in the body behind a marker), **error-based** (version in a reflected error), **boolean-based** (a syntax valid only on that engine returns the normal page — for fully blind points with no reflection), a **behavioral boolean** (a NoSQL `||'1'=='1'` TRUE/FALSE differential, which detects **MongoDB** even without a reflected error) and **time-based** (an engine-specific delay). Covers MySQL/MariaDB, Microsoft SQL Server, PostgreSQL, Oracle, SQLite, MongoDB, and best-effort GraphQL:

```bash
./dirrunner sqli -u 'https://TARGET/item?id=1' --fingerprint
# -> DBMS: MySQL/MariaDB 8.0.32 (parameter [id])
```

### SQLi Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. Include query parameters to test them. |
| `--url-file FILE` | File with target URLs, one per line. |
| `--method / -X METHOD` | HTTP method. Default `GET`; `POST` is implied when a body is set. |
| `--data / -p FORM` | Form-urlencoded body to test (e.g. `user=FUZZ&pass=x`). `FUZZ` marks the field, else all fields. |
| `--json-body / -b JSON` | JSON body to test. `FUZZ` marks the field, else all top-level string fields. |
| `--csrf-field NAME` | Refresh this anti-CSRF token (and its session cookie) before each request. |
| `--csrf-url URL` | Page to GET the fresh token from. Default: the target URL. |
| `--regx PATTERN` | Also report (and show) responses matching this regex/word — useful for spotting leaked/extracted data (a flag, a username, a version). |
| `--fingerprint` | Identify the DBMS engine/version instead of running the full scan. |
| `--payload-wordlist / -w FILE` | SQLi payload wordlist. Default `wordlist/sqli-payloads.txt`. |
| `--header-wordlist FILE` | Header-name wordlist to test. Default `wordlist/sqli-headers.txt`. |
| `--error-wordlist FILE` | SQL error-pattern wordlist (one regex per line). Default `wordlist/sqli-errors.txt`. |
| `--boolean-wordlist FILE` | Boolean TRUE/FALSE pairs, TAB separated. Default `wordlist/sqli-boolean.txt`. |
| `--sleep-wordlist FILE` | Time-based payload templates (`%d` = seconds). Default `wordlist/sqli-sleep.txt`. |
| `--encode LIST` | Chain encodings in order (`base64`, `urlencode`, `xor`); only the encoded payload is sent. |
| `--xor-key KEY` | XOR key used with `--encode xor`. Default is `DirRunner`. |
| `--sleep N` | Delay in seconds for time-based (SLEEP) detection. Default is `3`; `0` disables it. |
| `--continue-on-success` | Keep scanning a target after the first finding. |

### SQLi Examples

```bash
./dirrunner sqli -u 'https://example.com/item?id=1'
./dirrunner sqli -u 'https://example.com/item?id=1' --encode base64,urlencode
./dirrunner sqli --url-file targets.txt --continue-on-success
```

### SQLi Scenarios

The injection point and how to target it, with the technique that catches each one. Add `--continue-on-success` to keep scanning past the first finding, `-v` to see every probe, and `-j` for JSON output.

| Scenario | Command | How it's detected |
| --- | --- | --- |
| **Query parameter** (GET) | `sqli -u 'https://t/item?id=1'` | error / boolean / time / redirect, on every `?param` |
| **POST form**, test all fields | `sqli -u https://t/login -p 'username=admin&password=admin'` | each field injected in turn |
| **POST form**, one field | `sqli -u https://t/login -p 'username=FUZZ&password=admin'` | only the `FUZZ` field |
| **POST JSON** | `sqli -u https://t/api/login -b '{"user":"admin","pass":"x"}'` | each string field (and `FUZZ` marks one) |
| **CSRF-protected login** | `sqli -u https://t/login -X POST --data 'csrf=x&username=FUZZ&password=p' --csrf-field csrf` | fresh token+cookie fetched per request |
| **Blind, cookie** (conditional response) | `sqli -u https://t/ -C 'TrackingId=FUZZ'` | boolean TRUE/FALSE differential (token auto-fetched) |
| **Blind, header** | `sqli -u https://t/ -H 'X-Forwarded-For: <ip>FUZZ'` | same, on the header |
| **Login bypass** | `sqli -u https://t/login -p 'username=FUZZ&password=x'` | a payload that newly redirects (e.g. `administrator'--`) |
| **Blind, time-based** | `sqli -u 'https://t/filter?category=FUZZ' --sleep 5` | an engine-specific `SLEEP`/`pg_sleep`/`WAITFOR` delay |
| **NoSQL operator** (MongoDB, JSON login) | `sqli -u https://t/login -b '{"username":"admin","password":"admin"}'` | all string fields → operator objects, status/length change |
| **Identify the DBMS** | `sqli -u 'https://t/item?id=1' --fingerprint` | UNION / error / boolean / NoSQL operator / time, per engine |
| **Match leaked content** | `sqli -u 'https://t/item?id=1' --regx 'FLAG\{[^}]+\}'` | reports (and shows) any response matching the pattern |

Notes:
- Use the channel the app actually reads: a `TrackingId` **cookie** is `-C`, not `-H`; a JSON login is `-b`, not `--data` (a form body often returns `400`).
- For a blind cookie/header, put the real value **before** `FUZZ` (or rely on the auto-fetched cookie) so the query matches a row.

## Crawl Module

Use `crawl` to walk a site breadth-first and discover page URLs plus JavaScript and image assets. `--emails` also extracts e-mail addresses from each page, and `--regx` reports every match of a pattern. By default the crawl stays on the target's host.

When `--emails` and/or `--regx` are given, `crawl` runs in **extraction mode**: it still follows links internally to reach more pages, but it prints **only the findings** (each e-mail/match and the page it was found on), not the URL/JS/image inventory. Without those flags it prints the full discovery inventory as before.

`--emails` extraction uses a letters-only TLD with word boundaries and an asset-extension blocklist, so it no longer reports a trailing dot (`a@b.io.`), a word glued to the address, or retina-image names (`logo@2x.png`, `icon@3x.svg`) as e-mails; duplicates are collapsed case-insensitively.

### Crawl Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL. |
| `--url-file FILE` | File with target URLs, one per line. |
| `--max-depth / -D N` | Maximum crawl depth. Default is `3`. |
| `--emails` | Extract e-mail addresses from each page. |
| `--regx PATTERN` | Report every match of this regular expression (alias: `--regex`). |
| `--allow-external` | Follow links to other hosts (default: same host only). |

### Crawl Examples

```bash
./dirrunner crawl -u https://example.com --max-depth 2
./dirrunner crawl -u https://example.com --emails
./dirrunner crawl -u https://example.com --regx 'AKIA[0-9A-Z]{16}'
```

## Clone Module

Use `clone` to download a site. It follows same-domain links and assets (`a`, `link`, `script`, `img`) breadth-first with a pool of concurrent workers and writes each resource under `<project>/<host>/<path>`. Remote paths are sanitized so a response can never write outside the project directory.

### Clone Flags

| Flag | Description |
| --- | --- |
| `--url / -u URL` | Target URL to clone. |
| `--name / -n NAME` | Project directory name. Default is a generated name. |
| `--threads / -t N` | Concurrent download workers. Default is `10`. |

### Clone Examples

```bash
./dirrunner clone -u https://example.com -n example-copy
echo https://example.com | ./dirrunner clone -
```

## Recipes (use cases)

Practical one-liners for common scenarios. Add `-v` to see every request, `-j` for JSON, and `-o out.json` to export.

**Discover content on a host**
```bash
./dirrunner dir -u https://target.com -w wordlist/directory.txt -c 200,301,302,403
./dirrunner dir -u https://target.com -w wordlist/directory.txt -f wordlist/filename.txt -e php,txt,bak   # + files
./dirrunner dir -u https://target.com -w wordlist/directory.txt -r --depth 3                              # recursive
```

**Scan many hosts from a file (any URL module)**
```bash
./dirrunner dir --url-file hosts.txt -w wordlist/directory.txt -c 200
cat hosts.txt | ./dirrunner fingerprint - -j | jq -r '.[].url'
```

**Subdomains and virtual hosts**
```bash
./dirrunner dns -d target.com -w wordlist/subdomains.txt
./dirrunner vhost -u https://192.0.2.10 -d target.com -w wordlist/subdomains.txt
```

**Path traversal / LFI** (signature-based; add `--regx` for a custom marker)
```bash
./dirrunner traversal -u 'https://target.com/download?file=FUZZ'
./dirrunner traversal -u 'https://target.com/download?file=FUZZ' --encode base64 --continue-on-success
./dirrunner traversal -u 'https://target.com/?file=FUZZ' --regx 'DB_PASSWORD=\w+'
```

**401/403 bypass**
```bash
./dirrunner bypass -u https://target.com/admin --continue-on-success
```

**SQL injection — query parameter**
```bash
./dirrunner sqli -u 'https://target.com/item?id=1' --continue-on-success
```

**SQL injection — login form (POST)**
```bash
# test all fields
./dirrunner sqli -u https://target.com/login -p 'username=admin&password=admin' --continue-on-success
# focus one field with FUZZ
./dirrunner sqli -u https://target.com/login -p 'username=FUZZ&password=admin' --continue-on-success
```

**SQL injection — login form behind CSRF** (fetches a fresh token+cookie per request)
```bash
./dirrunner sqli -u https://target.com/login -X POST \
  --data 'csrf=x&username=FUZZ&password=admin' --csrf-field csrf --continue-on-success
```

**Blind SQLi in a cookie / header (conditional response)**
```bash
./dirrunner sqli -u https://target.com/ -C 'TrackingId=FUZZ' --continue-on-success   # token auto-fetched
```

**Identify the database engine/version**
```bash
./dirrunner sqli -u 'https://target.com/item?id=1' --fingerprint
./dirrunner sqli -u https://target.com/login -p 'username=FUZZ&password=x' --csrf-field csrf --fingerprint
```

**Credential fuzzing through a CSRF login** (fuzz also supports `--csrf-field`)
```bash
./dirrunner fuzz -u https://target.com/login -p 'csrf=x&username=admin&password=FUZZ' \
  --csrf-field csrf -w wordlist/passwords.txt -c 302
```

**Crawl / clone**
```bash
./dirrunner crawl -u https://target.com --emails --max-depth 2
./dirrunner crawl -u https://target.com --regx 'AKIA[0-9A-Z]{16}'
./dirrunner clone -u https://target.com -n target-copy
```

**Through a proxy (Burp) or Tor**
```bash
./dirrunner dir -u https://target.com -w wordlist/directory.txt --proxy http://127.0.0.1:8080
./dirrunner dir -u https://target.com -w wordlist/directory.txt --proxy socks5h://127.0.0.1:9050
```

**Filter results with jq**
```bash
./dirrunner sqli -u https://target.com/login -p 'username=FUZZ&password=x' --csrf-field csrf -j 2>/dev/null \
  | jq -r '.[] | select(.value | test("bypass")) | "\(.status)\t\(.value)"'
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
