from core.DirRunner.main import *
import validators
import sys
import requests
from pathlib import Path
from datetime import datetime
import concurrent.futures
import urllib3
urllib3.disable_warnings()

class FILE_OPTION:
	MODULE_NAME="File enumeration"
	TARGET_URL=""
	WORDLIST=""
	USER_AGENT=""
	THREADS=10
	OUTPUT=""
	EXTENSIONS=[]
	HELP=False
	NO_TLS_VALIDATION=True
	TIMEOUT=15
	COOKIE=""
	USERNAME=""
	PASSWORD=""

class FILE_MODULE:

	def main(args):
		print(TerminalColor.Green +'file enumeration mode selected'+TerminalColor.Reset)

		FILE_OPTION.TARGET_URL=args.url
		FILE_OPTION.WORDLIST=args.wordlist
		FILE_OPTION.USER_AGENT=args.user_agent
		FILE_OPTION.THREADS=args.threads
		FILE_OPTION.HELP=args.help
		FILE_OPTION.TIMEOUT=int(args.timeout)
		FILE_OPTION.NO_TLS_VALIDATION=args.no_tls_validation
		FILE_OPTION.COOKIE=args.cookie
		FILE_OPTION.USERNAME=args.username
		FILE_OPTION.password=args.password

		if FILE_OPTION.HELP==True:
			FILE_HELP.Help()

		if args.exts:
			FILE_OPTION.EXTENSIONS=args.exts.split(",")


		if args.output:
			now = datetime.now()
			FILE_OPTION.OUTPUT=f'{now}-{args.output}'
			FILE_OUTPUT.Reset()

		if not FILE_OPTION.TARGET_URL or not FILE_OPTION.EXTENSIONS:
			print(TerminalColor.Red +"target url and exts file are required!"+TerminalColor.Reset)
			print(f"{TerminalColor.Orange}example 'python3 DirRunner.py file -u https://www.domain.com -x txt,php'{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py file -h' for commands{TerminalColor.Reset}")
			sys.exit()
		else:
			if not FILE_OPTION.WORDLIST:
				FILE_OPTION.WORDLIST="wordlist/filename.txt"

			file = Path(FILE_OPTION.WORDLIST)
			if not file.is_file():
				print(TerminalColor.LightRed +f"file {FILE_OPTION.WORDLIST} not found!"+TerminalColor.Reset)
				sys.exit()

			Banner.DirRunnerBanner()
			FILE_MODULE.Banner()

			if not validators.url(FILE_OPTION.TARGET_URL):
				print(TerminalColor.Red +"Invalid url!"+TerminalColor.Reset)
				sys.exit()

			try:
				print(f'[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking connection for {FILE_OPTION.TARGET_URL}{TerminalColor.Reset}')
				headers={"User-Agent":f"{FILE_OPTION.USER_AGENT}","cookie":FILE_OPTION.COOKIE}

				res = requests.get(FILE_OPTION.TARGET_URL,allow_redirects=False,auth=(FILE_OPTION.USERNAME, FILE_OPTION.PASSWORD),headers=headers,timeout=FILE_OPTION.TIMEOUT,verify=FILE_OPTION.NO_TLS_VALIDATION)
				print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Timeout for {FILE_OPTION.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.SSLError:
				print(f"{TerminalColor.Red}SSL verification error! add -k arg to ignore.{FILE_OPTION.TARGET_URL}{TerminalColor.Reset}")
				print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py file -h' for commands{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.TooManyRedirects:
				print(f"{TerminalColor.Red}Too may redirect for {FILE_OPTION.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.ConnectionError as e:
				print(f"{TerminalColor.Red}Connection error: {e}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.RequestException as e:
				raise SystemExit(e)
				sys.exit()

			FILE_TASK.Threads()


	def Banner():
		Message=f"""- Target: {TerminalColor.Green}{FILE_OPTION.TARGET_URL}{TerminalColor.Reset}
- Attack mode: {TerminalColor.Green}{FILE_OPTION.MODULE_NAME}{TerminalColor.Reset}
- Threads: {TerminalColor.Green}{FILE_OPTION.THREADS}{TerminalColor.Reset}
- Extensions: {TerminalColor.Green}{FILE_OPTION.EXTENSIONS}{TerminalColor.Reset}
- User-agent: {TerminalColor.Green}{FILE_OPTION.USER_AGENT}{TerminalColor.Reset}
- Wordlist file: {TerminalColor.Green}{FILE_OPTION.WORDLIST}{TerminalColor.Reset}
- Timeout: {TerminalColor.Green}{FILE_OPTION.TIMEOUT}s{TerminalColor.Reset}"""

		if FILE_OPTION.COOKIE:
			Message=f"""{Message}
- Cookie: {TerminalColor.Green}{FILE_OPTION.COOKIE}{TerminalColor.Reset}"""

		if FILE_OPTION.USERNAME:
			Message=f"""{Message}
- Username (http auth): {TerminalColor.Green}{FILE_OPTION.USERNAME}{TerminalColor.Reset}"""

		if FILE_OPTION.NO_TLS_VALIDATION==False:
			Message=f"""{Message}
- TLS Validation: {TerminalColor.Green}{FILE_OPTION.NO_TLS_VALIDATION}{TerminalColor.Reset}"""
		
		print(f"""{Message}
======================================================================================================""")

class FILE_TASK:

	def Threads():
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(FILE_OPTION.THREADS)) as executor:
				f = open(FILE_OPTION.WORDLIST,'r')
				future_to_url = {executor.submit(FILE_TASK.Run,word): word for word in f.read().split("\n")}

				for future in concurrent.futures.as_completed(future_to_url):
					future.result()

		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}                              ')
			sys.exit()

	def Run(Line=""):

		headers={"User-Agent":f"{FILE_OPTION.USER_AGENT}","cookie":FILE_OPTION.COOKIE}

		READ_URL = FILE_OPTION.TARGET_URL[len(FILE_OPTION.TARGET_URL) - 1]
		if READ_URL =='/':
			URL_TO_REQUEST=f'{FILE_OPTION.TARGET_URL}{Line}'
		else:
			URL_TO_REQUEST=f'{FILE_OPTION.TARGET_URL}/{Line}'
		
		print(f'processing: {Line}                                                                 ',end="\r")

		for extension in FILE_OPTION.EXTENSIONS:
			BUILD_URL_TO_REQUEST=f'{URL_TO_REQUEST}.{extension.replace(".","")}'
			try:
				res = requests.get(BUILD_URL_TO_REQUEST,allow_redirects=False,auth=(FILE_OPTION.USERNAME, FILE_OPTION.PASSWORD),headers=headers,timeout=FILE_OPTION.TIMEOUT,verify=FILE_OPTION.NO_TLS_VALIDATION)

				if res.status_code==200:
					print(f'[{TerminalColor.Green}+{TerminalColor.Reset}] {TerminalColor.Green}{BUILD_URL_TO_REQUEST}{TerminalColor.Reset}                        ')
					FILE_OUTPUT.Run(BUILD_URL_TO_REQUEST)
			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Timeout for {BUILD_URL_TO_REQUEST}{TerminalColor.Reset}")
			except:
				return 0


class FILE_OUTPUT:

	def Run(MSG):
		if FILE_OPTION.OUTPUT:
			if '.html' in FILE_OPTION.OUTPUT:
				FILE_OUTPUT.HTML(MSG)
			else:
				FILE_OUTPUT.TXT(MSG)

	def TXT(MSG):
		with open(FILE_OPTION.OUTPUT, 'a+') as f:
			f.write(f"{MSG}\n")


	def HTML(MSG):
		contents=""
		with open(FILE_OPTION.OUTPUT, "r") as f:
			contents = f.readlines()

		text=f"""

<tr>
	<td>{MSG}</td>
</tr>

"""
		index=(len(contents) -6)

		contents.insert(index, text)

		with open(FILE_OPTION.OUTPUT, "w") as f:
			contents = "".join(contents)
			f.write(contents)


	def Reset():
		if '.html' in FILE_OPTION.OUTPUT:
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
			<h1 class="text-center">DirRunner v1.0 - File enumeration module</h1>
			<h3>Target URL: {FILE_OPTION.TARGET_URL}</h3>
			<hp>This report show all files detected on {FILE_OPTION.TARGET_URL} by File enumeration module</p>

			<table class="table">
   		 		<thead>
      				<tr>
        				<th>File path</th>
      				</tr>
    			</thead>
    			<tbody>



    			</tbody>
  			</table>
		</div>
	</body>
</html>"""

			with open(FILE_OPTION.OUTPUT, 'w') as f:
				f.write(header)
		else:
			with open(FILE_OPTION.OUTPUT, 'w') as f:
				f.write("")

class FILE_HELP:
		def Help():
			print("""File - Help menu

Uses file enumeration mode

Usage:
  python3 DirRunner.py file [args]

Args
	-u, --url                 set target url (required)
	-x, --exts                set exts files [txt,php] (required)
	-a, --user-agent          set user-agent 'DirRunner v1.0' by default
	-w, --wordlist            set wordlist file
	-t, --threads             set threads
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

	file enumeration
	use: python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x txt,php

	txt output
	use: python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x txt,php -o report.txt

	html output
	use: python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x txt,php -o report.html

				""")
			sys.exit()









