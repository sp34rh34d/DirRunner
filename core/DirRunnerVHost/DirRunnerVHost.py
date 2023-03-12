from core.DirRunner.main import *
import validators
import sys
import requests
from pathlib import Path
import concurrent.futures
from datetime import datetime
import urllib3
urllib3.disable_warnings()

class VHOST_OPTIONS:
	MODULE_NAME="Virtual host Enumeration"
	WORDLIST="wordlist/subdomains.txt"
	TARGET_URL=""
	TARGET_DOMAIN=""
	USER_AGENT=""
	THREADS=10
	OUTPUT=""
	NO_TLS_VALIDATION=True
	COOKIE=""
	HELP=False

class VHOST_MODULE:

	def main(args):
		print(TerminalColor.Green +'virtual host enumeration mode selected'+TerminalColor.Reset)

		VHOST_OPTIONS.WORDLIST=args.wordlist
		VHOST_OPTIONS.TARGET_DOMAIN=args.domain
		VHOST_OPTIONS.TARGET_URL=args.url
		VHOST_OPTIONS.THREADS=args.threads
		VHOST_OPTIONS.USER_AGENT=args.user_agent
		VHOST_OPTIONS.OUTPUT=args.output
		VHOST_OPTIONS.NO_TLS_VALIDATION=args.no_tls_validation
		VHOST_OPTIONS.HELP=args.help
		VHOST_OPTIONS.COOKIE=args.cookie

		if VHOST_OPTIONS.HELP==True:
			VHOST_HELP.Help()

		if args.output:
			now = datetime.now()
			VHOST_OPTIONS.OUTPUT=f'{now}-{args.output}'
			VHOST_OUTPUT.Reset()

		if not VHOST_OPTIONS.TARGET_DOMAIN or not VHOST_OPTIONS.TARGET_URL:
			print(TerminalColor.Red +"target domain and target url are required!"+TerminalColor.Reset)
			print(f"{TerminalColor.Orange}example 'python3 DirRunner.py vhost -d domain.com -u http://10.10.10.10'{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py vhost -h' for commands{TerminalColor.Reset}")
			sys.exit()
		else:
			if not VHOST_OPTIONS.WORDLIST:
				VHOST_OPTIONS.WORDLIST="wordlist/subdomains.txt"

			file = Path(VHOST_OPTIONS.WORDLIST)
			if not file.is_file():
				print(f"{TerminalColor.Red}file {VHOST_OPTIONS.WORDLIST} not found!{TerminalColor.Reset}")
				sys.exit()

			Banner.DirRunnerBanner()
			VHOST_MODULE.Banner()
			if not validators.domain(VHOST_OPTIONS.TARGET_DOMAIN):
				print(TerminalColor.LightRed +"Invalid domain!"+TerminalColor.Reset)
				sys.exit()
			if not validators.url(VHOST_OPTIONS.TARGET_URL):
				print(TerminalColor.LightRed +"Invalid url!"+TerminalColor.Reset)
				sys.exit()

			VHOST_TASKS.Threads()

	def Banner():
		Message=f"""- Target: {TerminalColor.Green}{VHOST_OPTIONS.TARGET_DOMAIN}{TerminalColor.Reset}
- Host: {TerminalColor.Green}{VHOST_OPTIONS.TARGET_URL}{TerminalColor.Reset}
- Attack mode: {TerminalColor.Green}{VHOST_OPTIONS.MODULE_NAME}{TerminalColor.Reset}
- User-agent: {TerminalColor.Green}{VHOST_OPTIONS.USER_AGENT}{TerminalColor.Reset}
- Threads: {TerminalColor.Green}{VHOST_OPTIONS.THREADS}{TerminalColor.Reset}
- Wordlist file: {TerminalColor.Green}{VHOST_OPTIONS.WORDLIST}{TerminalColor.Reset}"""

		if VHOST_OPTIONS.COOKIE:
			Message=f"""{Message}
- Cookie: {TerminalColor.Green}{VHOST_OPTIONS.COOKIE}{TerminalColor.Reset}"""

		if VHOST_OPTIONS.NO_TLS_VALIDATION==False:
			Message=f"""{Message}
- TLS Validation: {TerminalColor.Green}{VHOST_OPTIONS.NO_TLS_VALIDATION}{TerminalColor.Reset}"""
		
		print(f"""{Message}
======================================================================================================""")

class VHOST_TASKS:
	def Threads():
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(VHOST_OPTIONS.THREADS)) as executor:
				f = open(VHOST_OPTIONS.WORDLIST,'r')
				future_to_url = {executor.submit(VHOST_TASKS.Run,word): word for word in f.read().split("\n")}

				for future in concurrent.futures.as_completed(future_to_url):
					future.result()

		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}                              ')
			sys.exit()

	def Run(Line=""):

		DOMAIN_TO_REQUEST=f'{Line}.{VHOST_OPTIONS.TARGET_DOMAIN}'
		HOSTNAME=DOMAIN_TO_REQUEST.replace("..",".")
		headers={"User-Agent":f"{VHOST_OPTIONS.USER_AGENT}","cookie":VHOST_OPTIONS.COOKIE,"Host":HOSTNAME}

		print(f'processing: {HOSTNAME}                          ',end="\r")

		try:
			res = requests.get(VHOST_OPTIONS.TARGET_URL,headers=headers,allow_redirects=False,timeout=2,verify=VHOST_OPTIONS.NO_TLS_VALIDATION)
			VHOST_OUTPUT.ReadResponseCode(res,HOSTNAME)
	
		except requests.exceptions.SSLError:
			print(f"{TerminalColor.Red}SSL verification error! add -k arg to ignore.{VHOST_OPTIONS.TARGET_URL}{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py vhost -h' for commands{TerminalColor.Reset}")
			sys.exit()
		except:
			return 0


class VHOST_OUTPUT:
	def ReadResponseCode(ResponseHeaders="",MSG=""):

		if ResponseHeaders.status_code == 200:
			print(f'[{TerminalColor.Green}+{TerminalColor.Reset}] {TerminalColor.Green}{MSG}{TerminalColor.Reset}                                   ')
			VHOST_OUTPUT.Run(MSG)

	def Run(MSG=""):
		if VHOST_OPTIONS.OUTPUT:
			if '.html' in VHOST_OPTIONS.OUTPUT:
				VHOST_OUTPUT.HTML(MSG)
			else:
				VHOST_OUTPUT.TXT(MSG)

	def TXT(MSG=""):
		with open(VHOST_OPTIONS.OUTPUT, 'a+') as f:
			f.write(f"{MSG}\n")

	def HTML(MSG=""):
		contents=""
		with open(VHOST_OPTIONS.OUTPUT, "r") as f:
			contents = f.readlines()

		text=f"""

<tr>
	<td>{MSG}</td>
</tr>

"""
		index=(len(contents) -6)

		contents.insert(index, text)

		with open(VHOST_OPTIONS.OUTPUT, "w") as f:
			contents = "".join(contents)
			f.write(contents)


	def Reset():
		if '.html' in VHOST_OPTIONS.OUTPUT:
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
			<h1 class="text-center">DirRunner v1.0 - Virtual host enumeration module</h1>
			<h3>Target domain: {VHOST_OPTIONS.TARGET_DOMAIN}</h3>
			<hp>This report show all virtual hosts detected in Virtual host enumeration module for {VHOST_OPTIONS.TARGET_URL} host</p>

			<table class="table">
   		 		<thead>
      				<tr>
        				<th>Virtual Host</th>
      				</tr>
    			</thead>
    			<tbody>




    			</tbody>
  			</table>
		</div>
	</body>
</html>"""

			with open(VHOST_OPTIONS.OUTPUT, 'w') as f:
				f.write(header)
		else:
			with open(VHOST_OPTIONS.OUTPUT, 'w') as f:
				f.write("")


class VHOST_HELP:
		def Help():
			print("""VHOST - Help menu

Uses Virtual Host enumeration mode

Usage:
  python3 DirRunner.py vhost [args]

Args
	-d, --domain              set target domain (required)
	-u, --url                 set target host (required)
	-a, --user-agent          set user-agent 'DirRunner v1.0' by default
	-c, --cookie              set cookies to use for the requests
	-k, --no-tls-validation   skip TLS certificate verification
	-w, --wordlist            set wordlist file
	-t, --threads             set threads
	-h, --help                show this message

Generate outputs files
     -o,--output: set filename to save data,
                  txt format :  -o report.txt
                  html format : -o report.html

Examples:

	vhost enumeration
	use: python3 DirRunner.py vhost -d domain.com -u http://10.10.10.10 -w wordlist.txt

	vhost enumeration and exclude response size [4590]
	use: python3 DirRunner.py vhost -d domain.com -u http://10.10.10.10 -w wordlist.txt --fs 4590

	txt output
	use: python3 DirRunner.py dns -d domain.com -u http://10.10.10.10 -w wordlist.txt -o report.txt

	html output
	use: python3 DirRunner.py dns -d domain.com -u http://10.10.10.10 -w wordlist.txt -o report.html

				""")
			sys.exit()











