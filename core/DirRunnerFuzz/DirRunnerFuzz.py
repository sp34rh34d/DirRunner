from core.DirRunner.main import *
import validators
import sys
import requests
from pathlib import Path
import concurrent.futures
import string
import random
from datetime import datetime

class FUZZ_OPTION:
	MODULE_NAME="Fuzz enumeration"
	TARGET_URL=""
	METHODS=[]
	WORDLIST=""
	USER_AGENT=""
	THREADS=10
	STATUS_CODE=[]
	OUTPUT=""

class FUZZ_MODULE:
	def main(args):
		print(TerminalColor.Green +'fuzz attack mode selected'+TerminalColor.Reset)

		FUZZ_OPTION.TARGET_URL=args.url
		FUZZ_OPTION.WORDLIST=args.wordlist
		FUZZ_OPTION.USER_AGENT=args.user_agent
		FUZZ_OPTION.THREADS=args.threads
		FUZZ_OPTION.STATUS_CODE=args.status_code.split(",")
		FUZZ_OPTION.METHODS=args.method.split(",")

		if args.output:
			now = datetime.now()
			FUZZ_OPTION.OUTPUT=f'{now}-{args.output}'
			FUZZ_OUTPUT.Reset()

		if not FUZZ_OPTION.TARGET_URL:
			print(TerminalColor.Red +"target url is required!"+TerminalColor.Reset)
			print(f"{TerminalColor.Orange}example 'python3 DirRunner.py fuzz -u https://www.domain.com/FUZZ'{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py help' for commands{TerminalColor.Reset}")
			sys.exit()
		else:
			if not FUZZ_OPTION.WORDLIST:
				FUZZ_OPTION.WORDLIST="wordlist/directory.txt"

			file = Path(FUZZ_OPTION.WORDLIST)
			if not file.is_file():
				print(TerminalColor.LightRed +f"file {FUZZ_OPTION.WORDLIST} not found!"+TerminalColor.Reset)
				sys.exit()

			Banner.DirRunnerBanner()
			FUZZ_MODULE.Banner()

			if not validators.url(FUZZ_OPTION.TARGET_URL):
				print(TerminalColor.Red +"Invalid url!"+TerminalColor.Reset)
				sys.exit()

			if not "FUZZ" in FUZZ_OPTION.TARGET_URL:
				print(f"{TerminalColor.Red}FUZZ word not detected in url!{TerminalColor.Reset}")
				print(f"{TerminalColor.Orange}example 'python3 DirRunner.py fuzz -u https://www.domain.com/FUZZ'{TerminalColor.Reset}")
				print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py help' for commands{TerminalColor.Reset}")
				sys.exit()

			try:
				URL_TO_REQUEST = FUZZ_OPTION.TARGET_URL.replace("FUZZ","")
				print(f'[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking connection for {URL_TO_REQUEST}{TerminalColor.Reset}')
				headers={"User-Agent":f"{FUZZ_OPTION.USER_AGENT}"}
				res = requests.get(URL_TO_REQUEST,headers=headers,allow_redirects=False,timeout=5)
				print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Timeout for {URL_TO_REQUEST}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.ConnectionError as e:
				print(f"{TerminalColor.Red}Connection error: {e}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.TooManyRedirects:
				print(f"{TerminalColor.Red}Too may redirect for {URL_TO_REQUEST}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.RequestException as e:
				raise SystemExit(e)
				sys.exit()

			if "302" in FUZZ_OPTION.STATUS_CODE:
				if FUZZ_TASK.NonExistingUrlCheck(302):
					sys.exit()

			if "301" in FUZZ_OPTION.STATUS_CODE:
				if FUZZ_TASK.NonExistingUrlCheck(301):
					sys.exit()

			FUZZ_TASK.Threads()


	def Banner():
		print(f"""- Target: {TerminalColor.Green}{FUZZ_OPTION.TARGET_URL}{TerminalColor.Reset}
- Method: {TerminalColor.Green}{FUZZ_OPTION.METHODS}{TerminalColor.Reset}
- Attack mode: {TerminalColor.Green}{FUZZ_OPTION.MODULE_NAME}{TerminalColor.Reset}
- Threads: {TerminalColor.Green}{FUZZ_OPTION.THREADS}{TerminalColor.Reset}
- Status code: {TerminalColor.Green}{FUZZ_OPTION.STATUS_CODE}{TerminalColor.Reset}
- User-agent: {TerminalColor.Green}{FUZZ_OPTION.USER_AGENT}{TerminalColor.Reset}
- Wordlist file: {TerminalColor.Green}{FUZZ_OPTION.WORDLIST}{TerminalColor.Reset}
======================================================================================================""")

class FUZZ_TASK:

	def Threads():
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(FUZZ_OPTION.THREADS)) as executor:
				f = open(FUZZ_OPTION.WORDLIST,'r')
				future_to_url = {executor.submit(FUZZ_TASK.Run,word): word for word in f.read().split("\n")}

				for future in concurrent.futures.as_completed(future_to_url):
					future.result()

		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}                              ')
			sys.exit()

	def Run(Line=""):

		HEADERS={"User-Agent":f"{FUZZ_OPTION.USER_AGENT}"}

		URL_TO_REQUEST = FUZZ_OPTION.TARGET_URL.replace("FUZZ",Line)
		
		print(f'processing: {Line}                                                                 ',end="\r")

		try:
			if "GET" in FUZZ_OPTION.METHODS:
				res = requests.get(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
				FUZZ_OUTPUT.ReadResponseCode(res,Line,"GET",URL_TO_REQUEST)

			if "POST" in FUZZ_OPTION.METHODS:
				res = requests.post(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
				FUZZ_OUTPUT.ReadResponseCode(res,Line,"POST",URL_TO_REQUEST)

			if "PUT" in FUZZ_OPTION.METHODS:
				res = requests.put(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
				FUZZ_OUTPUT.ReadResponseCode(res,Line,"PUT",URL_TO_REQUEST)

			if "HEAD" in FUZZ_OPTION.METHODS:
				res = requests.head(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
				FUZZ_OUTPUT.ReadResponseCode(res,Line,"HEAD",URL_TO_REQUEST)

			if "DELETE" in FUZZ_OPTION.METHODS:
				res = requests.delete(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
				FUZZ_OUTPUT.ReadResponseCode(res,Line,"DELETE",URL_TO_REQUEST)

			if "OPTION" in FUZZ_OPTION.METHODS:
				res = requests.option(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
				FUZZ_OUTPUT.ReadResponseCode(res,Line,"OPTION",URL_TO_REQUEST)
	
		except requests.exceptions.Timeout:
			print(f"{TerminalColor.Red}Connection timeout for {URL_TO_REQUEST}{TerminalColor.Reset}")
		except:
			return 0

	def RandomStrings(size=28, chars=string.ascii_lowercase + string.digits):
		return ''.join(random.choice(chars) for _ in range(size))

	def NonExistingUrlCheck(Code):

		HEADERS={"User-Agent":f"{FUZZ_OPTION.USER_AGENT}"}

		RANDOM_URL=FUZZ_TASK.RandomStrings()
		URL_TO_REQUEST = FUZZ_OPTION.TARGET_URL.replace("FUZZ",RANDOM_URL)
		res=requests

		print(f"[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking for non existing urls response!{TerminalColor.Reset}",end="\r")
		print(f"[{TerminalColor.Blue}!{TerminalColor.Reset}] Testing random url {TerminalColor.Orange}{URL_TO_REQUEST}{TerminalColor.Reset} ...",end="\r")

		try:
			if "GET" in FUZZ_OPTION.METHODS:
				res = requests.get(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

			if "POST" in FUZZ_OPTION.METHODS:
				res = requests.post(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

			if "PUT" in FUZZ_OPTION.METHODS:
				res = requests.put(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

			if "HEAD" in FUZZ_OPTION.METHODS:
				res = requests.head(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

			if "DELETE" in FUZZ_OPTION.METHODS:
				res = requests.delete(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

			if "OPTION" in FUZZ_OPTION.METHODS:
				res = requests.option(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

			if res.status_code==Code:
				print(f"[{TerminalColor.Red}-{TerminalColor.Reset}] The website return a status code {TerminalColor.Orange}{res.status_code}{TerminalColor.Reset} for non existing urls {TerminalColor.Orange}{URL_TO_REQUEST}{TerminalColor.Reset}, please exclude this code from outputs. You can add following args '-s 200'")
				print(f"{TerminalColor.Orange}example 'python3 DirRunner.py dir -u https://www.domain.com -s 200'{TerminalColor.Reset}")
				return True

		except requests.exceptions.Timeout:
			print(f"{TerminalColor.Red}Connection timeout for {URL_TO_REQUEST}{TerminalColor.Reset}")
			return True
		except:
			return 0


class FUZZ_OUTPUT:

	def ReadResponseCode(ResponseHeaders="",Line="",Method="",MSG=""):

		if ResponseHeaders.status_code == 200 and "200" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Green}200{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                        ')
			FUZZ_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 201 and "201" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Green}201{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                         ')
			FUZZ_OUTPUT.Run(MSG,Method)

		if ResponseHeaders.status_code == 301 and "301" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Yellow}301{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset} -> [{TerminalColor.Blue}{ResponseHeaders.headers["Location"]}{TerminalColor.Reset}]')
			FUZZ_OUTPUT.Run(f'{MSG} -> {ResponseHeaders.headers["Location"]}',Method)

		if ResponseHeaders.status_code == 302 and "302" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Yellow}302{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset} -> [{TerminalColor.Blue}{ResponseHeaders.headers["Location"]}{TerminalColor.Reset}]')
			FUZZ_OUTPUT.Run(f'{MSG} -> {ResponseHeaders.headers["Location"]}',Method)
			
		if ResponseHeaders.status_code == 400 and "400" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}400{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			FUZZ_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 401 and "401" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}401{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			FUZZ_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 403 and "403" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}403{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			FUZZ_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 404 and "404" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}404{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			FUZZ_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 405 and "405" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Blue}405{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
			FUZZ_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 500 and "500" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Red}500{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                           ')
			FUZZ_OUTPUT.Run(MSG,Method)
		if ResponseHeaders.status_code == 503 and "503" in FUZZ_OPTION.STATUS_CODE:
			print(f'[{TerminalColor.Red}503{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                           ')
			FUZZ_OUTPUT.Run(MSG,Method)

	def Run(URL,Method):
		if FUZZ_OPTION.OUTPUT:
			if '.html' in FUZZ_OPTION.OUTPUT:
				FUZZ_OUTPUT.HTML(URL,Method)
			else:
				FUZZ_OUTPUT.TXT(URL,Method)

	def TXT(URL,Method):
		Message=f"Method: {Method} - URL: {URL}"
		with open(FUZZ_OPTION.OUTPUT, 'a+') as f:
			f.write(f"{Message}\n")


	def HTML(URL,Method):
		contents=""
		with open(FUZZ_OPTION.OUTPUT, "r") as f:
			contents = f.readlines()

		text=f"""

<tr>
	<td>{URL}</td>
    <td>{Method}</td>
</tr>

"""
		index=(len(contents) -6)

		contents.insert(index, text)
		with open(FUZZ_OPTION.OUTPUT, "w") as f:
		 	contents = "".join(contents)
		 	f.write(contents)


	def Reset():
		if '.html' in FUZZ_OPTION.OUTPUT:
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
			<h1 class="text-center">DirRunner v1.0 - Fuzz enumeration module</h1>
			<h3>Target URL: {FUZZ_OPTION.TARGET_URL}</h3>
			<hp>This report show all directories detected on {FUZZ_OPTION.TARGET_URL} by Fuzz enumeration module</p>

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

			with open(FUZZ_OPTION.OUTPUT, 'w') as f:
				f.write(header)
		else:
			with open(FUZZ_OPTION.OUTPUT, 'w') as f:
				f.write("")
