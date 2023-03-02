from core.DirRunner.main import *
import validators
import sys
import requests
from pathlib import Path
import concurrent.futures
import string
import random
from datetime import datetime
import urllib3
urllib3.disable_warnings()

class DIRECTORY_OPTION:
	MODULE_NAME="Directories enumeration"
	TARGET_URL=""
	METHODS=[]
	WORDLIST=""
	USER_AGENT=""
	THREADS=10
	STATUS_CODE=[]
	OUTPUT=""
	HELP=False
	NO_TLS_VALIDATION=True
	TIMEOUT=15
	COOKIE=""
	USERNAME=""
	PASSWORD=""

class DIRECTORY_MODULE:

	def main(args):
		print(TerminalColor.Green +'dir enumeration mode selected'+TerminalColor.Reset)

		DIRECTORY_OPTION.TARGET_URL=args.url
		DIRECTORY_OPTION.WORDLIST=args.wordlist
		DIRECTORY_OPTION.USER_AGENT=args.user_agent
		DIRECTORY_OPTION.THREADS=args.threads
		DIRECTORY_OPTION.STATUS_CODE=args.status_code.split(",")
		DIRECTORY_OPTION.METHODS=args.method.split(",")
		DIRECTORY_OPTION.HELP=args.help
		DIRECTORY_OPTION.TIMEOUT=int(args.timeout)
		DIRECTORY_OPTION.NO_TLS_VALIDATION=args.no_tls_validation
		DIRECTORY_OPTION.COOKIE=args.cookie
		DIRECTORY_OPTION.USERNAME=args.username
		DIRECTORY_OPTION.password=args.password

		if DIRECTORY_OPTION.HELP==True:
			DIRECTORY_HELP.Help()

		if args.output:
			now = datetime.now()
			DIRECTORY_OPTION.OUTPUT=f'{now}-{args.output}'
			DIRECTORY_OUTPUT.Reset()

		if not DIRECTORY_OPTION.TARGET_URL:
			print(TerminalColor.Red +"target url is required!"+TerminalColor.Reset)
			print(f"{TerminalColor.Orange}example 'python3 DirRunner.py dir -u https://www.domain.com'{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py dir -h' for commands{TerminalColor.Reset}")
			sys.exit()
		else:
			if not DIRECTORY_OPTION.WORDLIST:
				DIRECTORY_OPTION.WORDLIST="wordlist/directory.txt"

			file = Path(DIRECTORY_OPTION.WORDLIST)
			if not file.is_file():
				print(TerminalColor.LightRed +f"file {DIRECTORY_OPTION.WORDLIST} not found!"+TerminalColor.Reset)
				sys.exit()

			Banner.DirRunnerBanner()
			DIRECTORY_MODULE.Banner()

			if not validators.url(DIRECTORY_OPTION.TARGET_URL):
				print(TerminalColor.Red +"Invalid url!"+TerminalColor.Reset)
				sys.exit()

			try:
				print(f'[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking connection for {DIRECTORY_OPTION.TARGET_URL}{TerminalColor.Reset}')
				headers={"User-Agent":f"{DIRECTORY_OPTION.USER_AGENT}","cookie":DIRECTORY_OPTION.COOKIE}

				res = requests.get(DIRECTORY_OPTION.TARGET_URL,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)
				print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Timeout for {DIRECTORY_OPTION.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.SSLError:
				print(f"{TerminalColor.Red}SSL verification error! add -k arg to ignore.{DIRECTORY_OPTION.TARGET_URL}{TerminalColor.Reset}")
				print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py dir -h' for commands{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.TooManyRedirects:
				print(f"{TerminalColor.Red}Too may redirect for {DIRECTORY_OPTION.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.ConnectionError as e:
				print(f"{TerminalColor.Red}Connection error: {e}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.RequestException as e:
				raise SystemExit(e)
				sys.exit()

			for code in DIRECTORY_OPTION.STATUS_CODE:
				if DIRECTORY_TASK.NonExistingUrlCheck(int(code)):
					sys.exit()

			DIRECTORY_TASK.Threads()


	def Banner():
		Message=f"""- Target: {TerminalColor.Green}{DIRECTORY_OPTION.TARGET_URL}{TerminalColor.Reset}
- Method: {TerminalColor.Green}{DIRECTORY_OPTION.METHODS}{TerminalColor.Reset}
- Attack mode: {TerminalColor.Green}{DIRECTORY_OPTION.MODULE_NAME}{TerminalColor.Reset}
- Threads: {TerminalColor.Green}{DIRECTORY_OPTION.THREADS}{TerminalColor.Reset}
- Status code: {TerminalColor.Green}{DIRECTORY_OPTION.STATUS_CODE}{TerminalColor.Reset}
- User-agent: {TerminalColor.Green}{DIRECTORY_OPTION.USER_AGENT}{TerminalColor.Reset}
- Wordlist file: {TerminalColor.Green}{DIRECTORY_OPTION.WORDLIST}{TerminalColor.Reset}
- Timeout: {TerminalColor.Green}{DIRECTORY_OPTION.TIMEOUT}{TerminalColor.Reset}"""

		if DIRECTORY_OPTION.COOKIE:
			Message=f"""{Message}
- Cookie: {TerminalColor.Green}{DIRECTORY_OPTION.COOKIE}{TerminalColor.Reset}"""

		if DIRECTORY_OPTION.USERNAME:
			Message=f"""{Message}
- Username (http auth): {TerminalColor.Green}{DIRECTORY_OPTION.USERNAME}{TerminalColor.Reset}"""

		if DIRECTORY_OPTION.NO_TLS_VALIDATION==False:
			Message=f"""{Message}
- TLS Validation: {TerminalColor.Green}{DIRECTORY_OPTION.NO_TLS_VALIDATION}{TerminalColor.Reset}"""
		
		print(f"""{Message}
======================================================================================================""")

class DIRECTORY_TASK:

	def Threads():
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(DIRECTORY_OPTION.THREADS)) as executor:
				f = open(DIRECTORY_OPTION.WORDLIST,'r')
				future_to_url = {executor.submit(DIRECTORY_TASK.Run,word): word for word in f.read().split("\n")}

				for future in concurrent.futures.as_completed(future_to_url):
					future.result()

		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}                              ')
			sys.exit()

	def Run(Line=""):

		headers={"User-Agent":f"{DIRECTORY_OPTION.USER_AGENT}","cookie":DIRECTORY_OPTION.COOKIE}

		READ_URL = DIRECTORY_OPTION.TARGET_URL[len(DIRECTORY_OPTION.TARGET_URL) - 1]
		if READ_URL =='/':
			URL_TO_REQUEST=f'{DIRECTORY_OPTION.TARGET_URL}{Line}'
		else:
			URL_TO_REQUEST=f'{DIRECTORY_OPTION.TARGET_URL}/{Line}'
		
		print(f'processing: {Line}                                                                 ',end="\r")

		try:
			if "GET" in DIRECTORY_OPTION.METHODS:
				res = requests.get(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)
				DIRECTORY_OUTPUT.ReadResponseCode(res,Line,"GET",URL_TO_REQUEST)

			if "POST" in DIRECTORY_OPTION.METHODS:
				res = requests.post(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)
				DIRECTORY_OUTPUT.ReadResponseCode(res,Line,"POST",URL_TO_REQUEST)

			if "PUT" in DIRECTORY_OPTION.METHODS:
				res = requests.put(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)
				DIRECTORY_OUTPUT.ReadResponseCode(res,Line,"PUT",URL_TO_REQUEST)

			if "HEAD" in DIRECTORY_OPTION.METHODS:
				res = requests.head(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)
				DIRECTORY_OUTPUT.ReadResponseCode(res,Line,"HEAD",URL_TO_REQUEST)

			if "DELETE" in DIRECTORY_OPTION.METHODS:
				res = requests.delete(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)
				DIRECTORY_OUTPUT.ReadResponseCode(res,Line,"DELETE",URL_TO_REQUEST)

			if "OPTION" in DIRECTORY_OPTION.METHODS:
				res = requests.option(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)
				DIRECTORY_OUTPUT.ReadResponseCode(res,Line,"OPTION",URL_TO_REQUEST)
	
		except requests.exceptions.Timeout:
			print(f"{TerminalColor.Red}Connection timeout for {URL_TO_REQUEST}{TerminalColor.Reset}")
		except:
			return 0

	def RandomStrings(size=28, chars=string.ascii_lowercase + string.digits):
		return ''.join(random.choice(chars) for _ in range(size))

	def NonExistingUrlCheck(Code):

		headers={"User-Agent":f"{DIRECTORY_OPTION.USER_AGENT}","cookie":DIRECTORY_OPTION.COOKIE}

		RANDOM_URL=DIRECTORY_TASK.RandomStrings()
		READ_URL = DIRECTORY_OPTION.TARGET_URL[len(DIRECTORY_OPTION.TARGET_URL) - 1]
		res=requests

		if READ_URL =='/':
			URL_TO_REQUEST=f'{DIRECTORY_OPTION.TARGET_URL}{RANDOM_URL}'
		else:
			URL_TO_REQUEST=f'{DIRECTORY_OPTION.TARGET_URL}/{RANDOM_URL}'

		print(f"[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking for non existing urls response!{TerminalColor.Reset}",end="\r")
		print(f"[{TerminalColor.Blue}!{TerminalColor.Reset}] Testing random url {TerminalColor.Orange}{URL_TO_REQUEST}{TerminalColor.Reset} ...",end="\r")

		try:
			if "GET" in DIRECTORY_OPTION.METHODS:
				res = requests.get(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)

			if "POST" in DIRECTORY_OPTION.METHODS:
				res = requests.post(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)

			if "PUT" in DIRECTORY_OPTION.METHODS:
				res = requests.put(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)

			if "HEAD" in DIRECTORY_OPTION.METHODS:
				res = requests.head(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION0)

			if "DELETE" in DIRECTORY_OPTION.METHODS:
				res = requests.delete(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)

			if "OPTION" in DIRECTORY_OPTION.METHODS:
				res = requests.option(URL_TO_REQUEST,auth=(DIRECTORY_OPTION.USERNAME, DIRECTORY_OPTION.PASSWORD),headers=headers,allow_redirects=False,timeout=DIRECTORY_OPTION.TIMEOUT,verify=DIRECTORY_OPTION.NO_TLS_VALIDATION)

			if res.status_code==Code:
				print(f"[{TerminalColor.Red}-{TerminalColor.Reset}] The website return a status code {TerminalColor.Orange}{res.status_code}{TerminalColor.Reset} for non existing urls {TerminalColor.Orange}{URL_TO_REQUEST}{TerminalColor.Reset}, please exclude this code from outputs.")
				print(f"{TerminalColor.Orange}type 'python3 DirRunner.py dir -h' for commands{TerminalColor.Reset}")
				return True

		except requests.exceptions.Timeout:
			print(f"{TerminalColor.Red}Connection timeout for {URL_TO_REQUEST}{TerminalColor.Reset}")
			return True
		except:
			return 0


class DIRECTORY_OUTPUT:

	def ReadResponseCode(ResponseHeaders="",Line="",Method="",MSG=""):

		if ResponseHeaders.status_code == 200 and "200" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Green}200{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                        ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 201 and "201" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Green}201{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                         ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		
		if ResponseHeaders.status_code == 301 and "301" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Yellow}301{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset} -> [{TerminalColor.Blue}{ResponseHeaders.headers["Location"]}{TerminalColor.Reset}]')
			DIRECTORY_OUTPUT.Run(f'{MSG} -> {ResponseHeaders.headers["Location"]}',Method)
		
		if ResponseHeaders.status_code == 302 and "302" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Yellow}302{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset} -> [{TerminalColor.Blue}{ResponseHeaders.headers["Location"]}{TerminalColor.Reset}]')
			DIRECTORY_OUTPUT.Run(f'{MSG} -> {ResponseHeaders.headers["Location"]}',Method)
		
		if ResponseHeaders.status_code == 400 and "400" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}400{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 401 and "401" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}401{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 403 and "403" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}403{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 404 and "404" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}404{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 405 and "405" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}405{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 500 and "500" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Red}500{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                           ')
			DIRECTORY_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 503 and "503" in DIRECTORY_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Red}503{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                           ')
			DIRECTORY_OUTPUT.Run(MSG,Method)

	def Run(URL,Method):
		if DIRECTORY_OPTION.OUTPUT:
			if '.html' in DIRECTORY_OPTION.OUTPUT:
				DIRECTORY_OUTPUT.HTML(URL,Method)
			else:
				DIRECTORY_OUTPUT.TXT(URL,Method)

	def TXT(URL,Method):
		Message=f"Method: {Method} - URL: {URL}"
		with open(DIRECTORY_OPTION.OUTPUT, 'a+') as f:
			f.write(f"{Message}\n")


	def HTML(URL,Method):
		contents=""
		with open(DIRECTORY_OPTION.OUTPUT, "r") as f:
			contents = f.readlines()

		text=f"""

<tr>
	<td>{URL}</td>
    <td>{Method}</td>
</tr>

"""
		index=(len(contents) -6)

		contents.insert(index, text)

		with open(DIRECTORY_OPTION.OUTPUT, "w") as f:
			contents = "".join(contents)
			f.write(contents)


	def Reset():
		if '.html' in DIRECTORY_OPTION.OUTPUT:
			header=f""" 
<html>
	<head>
		<title>DirRunner report</title>
  		<meta charset="utf-8">
  		<meta name="viewport" content="width=device-width, initial-scale=1">
  		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/css/bootstrap.min.css">
  		<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.6.3/jquery.min.js"></script>
  		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.4.1/js/bootstrap.min.js"></script>
		
	</head>
	<body>
		<div class="container">
			<h1 class="text-center">DirRunner v1.0 - Directories enumeration module</h1>
			<h3>Target URL: {DIRECTORY_OPTION.TARGET_URL}</h3>
			<hp>This report show all directories detected on {DIRECTORY_OPTION.TARGET_URL} by Directories enumeration module</p>

			<table class="table">
   		 		<thead>
      				<tr>
        				<th>URL</th>
        				<th>Method</th>
      				</tr>
    			</thead>
    			<tbody>



    			</tbody>
  			</table>
		</div>
	</body>
</html>"""

			with open(DIRECTORY_OPTION.OUTPUT, 'w') as f:
				f.write(header)
		else:
			with open(DIRECTORY_OPTION.OUTPUT, 'w') as f:
				f.write("")

class DIRECTORY_HELP:
		def Help():
			print("""Directory - Help menu

Uses directory enumeration mode

Usage:
  python3 DirRunner.py dir [args]

Args
	-u, --url                 set target url (required)
	-a, --user-agent          set user-agent 'DirRunner v1.0' by default
	-s, --status-code         set the status code to print (200,301)
	-w, --wordlist            set wordlist file
	-t, --threads             set threads
	-m, --method              set method (GET/POST/DELETE/OPTION/PUT/HEAD) for HTTP requests, GET by default.
	-h, --help                show this message
	-c, --cookie              set cookies to use for the requests
	-k, --no-tls-validation   skip TLS certificate verification
	-P, --password            Password for Basic Auth
	-U, --username            Username for Basic Auth
	    --timeout             HTTP Timeout (default 15s)

Generate outputs files
     -o,--output: set filename to save data,
                  txt format :  -o report.txt
                  html format : -o report.html

Examples:

	dir enumeration
	use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt

	print only status code 200 and 301
	use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -s 200,301

	txt output
	use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -o report.txt

	html output
	use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -o report.html

				""")
			sys.exit()








