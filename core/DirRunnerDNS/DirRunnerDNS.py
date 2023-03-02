from core.DirRunner.main import *
import concurrent.futures
import validators
from pathlib import Path
import sys
import socket
from datetime import datetime

class DNS_OPTIONS:
	MODULE_NAME="DNS Enumeration"
	WORDLIST="wordlist/subdomains.txt"
	TARGET_DOMAIN=""
	THREADS=10
	OUTPUT=""
	HELP=False

class DNS_MODULE:

	def main(args):
		print(TerminalColor.Green +'dns attack mode selected'+TerminalColor.Reset)

		DNS_OPTIONS.WORDLIST=args.wordlist
		DNS_OPTIONS.TARGET_DOMAIN=args.domain
		DNS_OPTIONS.THREADS=args.threads
		DNS_OPTIONS.HELP=args.help

		if DNS_OPTIONS.HELP==True:
			DNS_HELP.Help()

		if args.output:
			now = datetime.now()
			DNS_OPTIONS.OUTPUT=f'{now}-{args.output}'
			DNS_OUTPUT.Reset()

		if not DNS_OPTIONS.TARGET_DOMAIN:
			print(TerminalColor.Red +"target domain is required!"+TerminalColor.Reset)
			print(f"{TerminalColor.Orange}example 'python3 DirRunner.py dns -d domain.com'{TerminalColor.Reset}")
			print(f"{TerminalColor.Orange}Type 'python3 DirRunner.py dns -h' for commands{TerminalColor.Reset}")
			sys.exit()
		else:
			if not DNS_OPTIONS.WORDLIST:
				DNS_OPTIONS.WORDLIST="wordlist/subdomains.txt"

			file = Path(DNS_OPTIONS.WORDLIST)
			if not file.is_file():
				print(f"{TerminalColor.Red}file {DNS_OPTIONS.WORDLIST} not found!{TerminalColor.Reset}")
				sys.exit()

			Banner.DirRunnerBanner()
			DNS_MODULE.Banner()
			if not validators.domain(DNS_OPTIONS.TARGET_DOMAIN):
				print(TerminalColor.LightRed +"Invalid domain!"+TerminalColor.Reset)
				sys.exit()

			DNS_TASKS.Threads()

	def Banner():
		print(f"""- Target: {TerminalColor.Green}{DNS_OPTIONS.TARGET_DOMAIN}{TerminalColor.Reset}
- Attack mode: {TerminalColor.Green}{DNS_OPTIONS.MODULE_NAME}{TerminalColor.Reset}
- Threads: {TerminalColor.Green}{DNS_OPTIONS.THREADS}{TerminalColor.Reset}
- Wordlist file: {TerminalColor.Green}{DNS_OPTIONS.WORDLIST}{TerminalColor.Reset}
======================================================================================================""")

class DNS_TASKS:

	def Threads():
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(DNS_OPTIONS.THREADS)) as executor:
				f = open(DNS_OPTIONS.WORDLIST,'r')
				future_to_url = {executor.submit(DNS_TASKS.Run,word): word for word in f.read().split("\n")}
		
				for future in concurrent.futures.as_completed(future_to_url):
					future.result()

		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}                              ')
			sys.exit()

	def Run(Line=""):
		try:

			DOMAIN_TO_REQUEST=f'{Line}.{DNS_OPTIONS.TARGET_DOMAIN}'
			HOSTNAME=DOMAIN_TO_REQUEST.replace("..",".")
			print(f'processing: {HOSTNAME}                          ',end="\r")
			res=socket.gethostbyname(HOSTNAME)
			print(f'[{TerminalColor.Green}+{TerminalColor.Reset}] {TerminalColor.Green}{HOSTNAME}{TerminalColor.Reset} - [{TerminalColor.Yellow}{res}{TerminalColor.Reset}]                                   ')
			DNS_OUTPUT.Run(HOSTNAME,res)

		except:
			return 0

class DNS_OUTPUT:

	def Run(MSG="",IP=""):
		if DNS_OPTIONS.OUTPUT:
			if '.html' in DNS_OPTIONS.OUTPUT:
				DNS_OUTPUT.HTML(MSG,IP)
			else:
				DNS_OUTPUT.TXT(MSG)

	def TXT(MSG=""):
		with open(DNS_OPTIONS.OUTPUT, 'a+') as f:
			f.write(f"{MSG}\n")

	def HTML(MSG="",IP=""):
		contents=""
		with open(DNS_OPTIONS.OUTPUT, "r") as f:
			contents = f.readlines()

		text=f"""

<tr>
	<td>{MSG}</td>
    <td>{IP}</td>
</tr>

"""
		index=(len(contents) -6)

		contents.insert(index, text)

		with open(DNS_OPTIONS.OUTPUT, "w") as f:
			contents = "".join(contents)
			f.write(contents)


	def Reset():
		if '.html' in DNS_OPTIONS.OUTPUT:
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
			<h1 class="text-center">DirRunner v1.0 - DNS enumeration module</h1>
			<h3>Target domain: {DNS_OPTIONS.TARGET_DOMAIN}</h3>
			<hp>This report show all subdomains and DNS (A) registers detected in DNS enumeration module</p>

			<table class="table">
   		 		<thead>
      				<tr>
        				<th>Hostname</th>
        				<th>IP</th>
      				</tr>
    			</thead>
    			<tbody>




    			</tbody>
  			</table>
		</div>
	</body>
</html>"""

			with open(DNS_OPTIONS.OUTPUT, 'w') as f:
				f.write(header)
		else:
			with open(DNS_OPTIONS.OUTPUT, 'w') as f:
				f.write("")



class DNS_HELP:
		def Help():
			print("""DNS - Help menu

Uses DNS enumeration mode

Usage:
  python3 DirRunner.py dns [args]

Args
	-d, --domain        set target domain (required)
	-w, --wordlist      set wordlist file
	-t, --threads       set threads
	-h, --help          show this message

Generate outputs files
     -o,--output: set filename to save data,
                  txt format :  -o report.txt
                  html format : -o report.html

Examples:

	dns enumeration
	use: python3 DirRunner.py dns -d domain.com -w wordlist.txt

	txt output
	use: python3 DirRunner.py dns -d domain.com -w wordlist.txt -o report.txt

	html output
	use: python3 DirRunner.py dns -d domain.com -w wordlist.txt -o report.html

				""")
			sys.exit()














