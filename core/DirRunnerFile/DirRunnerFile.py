from core.DirRunner.main import *
import validators
import sys
import requests
from pathlib import Path
from datetime import datetime
import concurrent.futures

class FILE_OPTION:
	MODULE_NAME="File enumeration"
	TARGET_URL=""
	WORDLIST=""
	USER_AGENT=""
	THREADS=10
	OUTPUT=""
	EXTENSIONS=[]

class FILE_MODULE:

	def main(args):
		print(TerminalColor.Green +'file enumeration mode selected'+TerminalColor.Reset)

		FILE_OPTION.TARGET_URL=args.url
		FILE_OPTION.WORDLIST=args.wordlist
		FILE_OPTION.USER_AGENT=args.user_agent
		FILE_OPTION.THREADS=args.threads

		if args.exts:
			FILE_OPTION.EXTENSIONS=args.exts.split(",")


		if args.output:
			now = datetime.now()
			FILE_OPTION.OUTPUT=f'{now}-{args.output}'
			FILE_OUTPUT.Reset()

		if not FILE_OPTION.TARGET_URL or not FILE_OPTION.EXTENSIONS:
			print(TerminalColor.Red +"target url and exts file are required!"+TerminalColor.Reset)
			print(f"{TerminalColor.Orange}example 'python3 DirRunner.py file -u https://www.domain.com -x txt,php'{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py help' for commands{TerminalColor.Reset}")
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
				headers={"User-Agent":f"{FILE_OPTION.USER_AGENT}"}
				res = requests.get(FILE_OPTION.TARGET_URL,headers=headers,allow_redirects=False,timeout=5)
				print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Timeout for {FILE_OPTION.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.ConnectionError as e:
				print(f"{TerminalColor.Red}Connection error: {e}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.TooManyRedirects:
				print(f"{TerminalColor.Red}Too may redirect for {FILE_OPTION.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.RequestException as e:
				raise SystemExit(e)
				sys.exit()

			FILE_TASK.Threads()


	def Banner():
		print(f"""- Target: {TerminalColor.Green}{FILE_OPTION.TARGET_URL}{TerminalColor.Reset}
- Attack mode: {TerminalColor.Green}{FILE_OPTION.MODULE_NAME}{TerminalColor.Reset}
- Threads: {TerminalColor.Green}{FILE_OPTION.THREADS}{TerminalColor.Reset}
- Extensions: {TerminalColor.Green}{FILE_OPTION.EXTENSIONS}{TerminalColor.Reset}
- User-agent: {TerminalColor.Green}{FILE_OPTION.USER_AGENT}{TerminalColor.Reset}
- Wordlist file: {TerminalColor.Green}{FILE_OPTION.WORDLIST}{TerminalColor.Reset}
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

		HEADERS={"User-Agent":f"{FILE_OPTION.USER_AGENT}"}

		READ_URL = FILE_OPTION.TARGET_URL[len(FILE_OPTION.TARGET_URL) - 1]
		if READ_URL =='/':
			URL_TO_REQUEST=f'{FILE_OPTION.TARGET_URL}{Line}'
		else:
			URL_TO_REQUEST=f'{FILE_OPTION.TARGET_URL}/{Line}'
		
		print(f'processing: {Line}                                                                 ',end="\r")

		for extension in FILE_OPTION.EXTENSIONS:
			BUILD_URL_TO_REQUEST=f'{URL_TO_REQUEST}.{extension.replace(".","")}'
			try:
				res = requests.get(BUILD_URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=15)

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









