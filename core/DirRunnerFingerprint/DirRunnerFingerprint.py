from core.DirRunner.main import *
import validators
import sys
import requests
from builtwith import *
from datetime import datetime

class FINGERPRINT_OPTIONS:
	MODULE_NAME="Fingerprint"
	TARGET_URL=""
	USER_AGENT="DirRunner v1.0"
	OUTPUT=""

class FINGERPRINT_MODULE:

	def main(args):
		print(TerminalColor.Green +'fingerprint mode selected'+TerminalColor.Reset)

		FINGERPRINT_OPTIONS.TARGET_URL=args.url
		FINGERPRINT_OPTIONS.USER_AGENT=args.user_agent

		if args.output:
			now = datetime.now()
			FINGERPRINT_OPTIONS.OUTPUT=f'{now}-{args.output}'
			FINGERPRINT_OUTPUT.Reset()

		if not FINGERPRINT_OPTIONS.TARGET_URL:
			print(TerminalColor.Red +"target url is required!"+TerminalColor.Reset)
			print(f"{TerminalColor.Orange}example 'python3 DirRunner.py fingerprint -u https://www.domain.com'{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py help' for commands{TerminalColor.Reset}")
			sys.exit()
		else:
			Banner.DirRunnerBanner()
			FINGERPRINT_MODULE.Banner()

			if not validators.url(FINGERPRINT_OPTIONS.TARGET_URL):
				print(TerminalColor.Red +"Invalid url!"+TerminalColor.Reset)
				sys.exit()

			try:
				print(f'[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking connection for {FINGERPRINT_OPTIONS.TARGET_URL}{TerminalColor.Reset}')
				headers={"User-Agent":f"{FINGERPRINT_OPTIONS.USER_AGENT}"}
				res = requests.get(FINGERPRINT_OPTIONS.TARGET_URL,headers=headers,allow_redirects=False,timeout=5)
				print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Timeout for {FINGERPRINT_OPTIONS.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.ConnectionError as e:
				print(f"{TerminalColor.Red}Connection error: {e}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.TooManyRedirects:
				print(f"{TerminalColor.Red}Too may redirect for {FINGERPRINT_OPTIONS.TARGET_URL}{TerminalColor.Reset}")
				sys.exit()
			except requests.exceptions.RequestException as e:
				raise SystemExit(e)
				sys.exit()

			FINGERPRINT_TASK.Run()


	def Banner():
		print(f"""- Target: {TerminalColor.Green}{FINGERPRINT_OPTIONS.TARGET_URL}{TerminalColor.Reset}
- Attack mode: {TerminalColor.Green}Fingerprint{TerminalColor.Reset}
======================================================================================================""")

class FINGERPRINT_TASK:

	def Run():
		
		headers={"User-Agent":f"{FINGERPRINT_OPTIONS.USER_AGENT}"}
		
		try:
			res = requests.get(FINGERPRINT_OPTIONS.TARGET_URL,headers=headers,timeout=15)
		except requests.exceptions.Timeout:
	 		print(f"{TerminalColor.Red}Timeout for {FINGERPRINT_OPTIONS.TARGET_URL}{TerminalColor.Reset}")
		except requests.exceptions.ConnectionError:
			print(f"{TerminalColor.Red}Connection error for {FINGERPRINT_OPTIONS.TARGET_URL}{TerminalColor.Reset}")
		except requests.exceptions.TooManyRedirects:
	 		print(f"{TerminalColor.Red}Too may redirect for {FINGERPRINT_OPTIONS.TARGET_URL}{TerminalColor.Reset}")

		try:
			print(f"Status: {TerminalColor.Green}{res.status_code}{TerminalColor.Reset}")
			print(f"Web server: {TerminalColor.Green}{res.headers['Server']}{TerminalColor.Reset}")
			print(f"Content-Length: {TerminalColor.Green}{res.headers['Content-Length']}{TerminalColor.Reset}")
		except:
			print("",end="\r")
	
		try:
			print(f"{TerminalColor.Orange}...wait...{TerminalColor.Reset}")
			tech=builtwith(FINGERPRINT_OPTIONS.TARGET_URL)
			for f in tech:
				print(f'{f} : {TerminalColor.Green}{tech[f]}{TerminalColor.Reset}')

			print(f"{TerminalColor.Green}Done!{TerminalColor.Reset}")
			FINGERPRINT_OUTPUT.Run(tech)
		except:
			print(f'{TerminalColor.Red}Error trying to get web tech!{TerminalColor.Reset}')


class FINGERPRINT_OUTPUT:

	def Run(MSG):
		if FINGERPRINT_OPTIONS.OUTPUT:
			if '.html' in FINGERPRINT_OPTIONS.OUTPUT:
				FINGERPRINT_OUTPUT.HTML(MSG)
			else:
				FINGERPRINT_OUTPUT.TXT(MSG)

	def TXT(MSG):
		with open(FINGERPRINT_OPTIONS.OUTPUT, 'a+') as f:
			for text in MSG:
				message=f"{text} : {MSG[text]}"
				f.write(f"{message}\n")

	def HTML(MSG):
		contents=""
		with open(FINGERPRINT_OPTIONS.OUTPUT, "r") as f:
			contents = f.readlines()

		message=""
		for text in MSG:
			message=f"""{message}
<tr>
	<td>{text}</td>
	<td>{MSG[text]}</td>
</tr>

"""
		index=(len(contents) -6)

		contents.insert(index, message)

		with open(FINGERPRINT_OPTIONS.OUTPUT, "w") as f:
			contents = "".join(contents)
			f.write(contents)


	def Reset():
		if '.html' in FINGERPRINT_OPTIONS.OUTPUT:
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
			<h1 class="text-center">DirRunner v1.0 - Fingerprint module</h1>
			<h3>Target url: {FINGERPRINT_OPTIONS.TARGET_URL}</h3>
			<hp>This report show web tecnology used for {FINGERPRINT_OPTIONS.TARGET_URL}</p>

			<table class="table">
   		 		<thead>
      				<tr>
        				<th>Name</th>
        				<th>Value</th>
      				</tr>
    			</thead>
    			<tbody>



    			</tbody>
  			</table>
		</div>
	</body>
</html>"""

			with open(FINGERPRINT_OPTIONS.OUTPUT, 'w') as f:
				f.write(header)
		else:
			with open(FINGERPRINT_OPTIONS.OUTPUT, 'w') as f:
				f.write("")









