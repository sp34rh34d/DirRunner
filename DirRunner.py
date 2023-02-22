#!/usr/bin/env python3

import requests
import argparse
import sys
import validators
from pathlib import Path
from core.dirRunnerMessage import *
from core.dirRunnerRequests import *

TARGET_URL=""
WORDLIST_FILE=""
STATUS_CODE=[]
EXTENSIONS=[]
THREADS=10
TARGET_DOMAIN=""
METHOD=[]
ATTACK_MODE=""
USER_AGENT=""

def DirectoryModule(URL="",Method=[],Wordlist="",Codes=[],UserAgent="",Thread=10):

	print(TerminalColor.Green + 'dir attack mode selected'+TerminalColor.Reset)

	if not URL:
		print(TerminalColor.LightRed +"some args are empty!"+TerminalColor.Reset)
		Banner.Help()
		sys.exit()
	else:
		if not Wordlist:
			Wordlist="wordlist/directory.txt"

		file = Path(Wordlist)
		if not file.is_file():
			print(TerminalColor.LightRed +f"file {Wordlist} not found!"+TerminalColor.Reset)
			sys.exit()

		Banner.DirRunnerBanner()
		Banner.DirModeBanner(URL,Method,Thread,Codes,UserAgent,Wordlist)
		if not validators.url(URL):
			print(TerminalColor.LightRed +"Invalid url!"+TerminalColor.Reset)
			sys.exit()

		try:
			print(f'[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking connection for {URL}{TerminalColor.Reset}')
			headers={"User-Agent":f"{UserAgent}"}
			res = requests.get(URL,headers=headers,allow_redirects=False,timeout=5)
			print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

		except requests.exceptions.Timeout:
			print(f"{TerminalColor.Red}Timeout for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.ConnectionError:
			print(f"{TerminalColor.Red}Connection error for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.TooManyRedirects:
			print(f"{TerminalColor.Red}Too may redirect for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.RequestException as e:
			raise SystemExit(e)
			sys.exit()

		if ResponseCode.ResponseCode_302==True:
			if Modules.NonExistingUrlCheck(URL,Method,UserAgent):
				sys.exit()

		Tasks.DirModuleThreads(URL,Method,Wordlist,UserAgent,Thread)


def FileModule(URL="",Wordlist="",Codes=[],Exts=[],UserAgent="",Thread=10):
	print(TerminalColor.Green +'file discovery mode selected'+TerminalColor.Reset)

	if not URL or not Exts:
		print(TerminalColor.LightRed +"some args are empty!"+TerminalColor.Reset)
		Banner.Help()
		sys.exit()
	else:
		if not Wordlist:
			Wordlist="wordlist/filename.txt"

		file = Path(Wordlist)
		if not file.is_file():
			print(TerminalColor.LightRed +f"file {Wordlist} not found!"+TerminalColor.Reset)
			sys.exit()

		Banner.DirRunnerBanner()
		Banner.FileModeBanner(URL,Thread,Codes,UserAgent,Exts,Wordlist)
		if not validators.url(URL):
			print(TerminalColor.LightRed +"Invalid url!"+TerminalColor.Reset)
			sys.exit()

		try:
			print(f'[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking connection for {URL}{TerminalColor.Reset}')
			headers={"User-Agent":f"{UserAgent}"}
			res = requests.get(URL,headers=headers,allow_redirects=False,timeout=5)
			print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

		except requests.exceptions.Timeout:
			print(f"{TerminalColor.Red}Timeout for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.ConnectionError:
			print(f"{TerminalColor.Red}Connection error for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.TooManyRedirects:
			print(f"{TerminalColor.Red}Too may redirect for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.RequestException as e:
			raise SystemExit(e)
			sys.exit()

		Tasks.FileModuleThreads(URL,Wordlist,UserAgent,Exts,Thread)


def FingerprintModule(URL="",UserAgent=""):
	print(TerminalColor.Green +'fingerprint mode selected'+TerminalColor.Reset)

	if not URL:
		print(TerminalColor.LightRed +"some args are empty!"+TerminalColor.Reset)
		Banner.Help()
		sys.exit()
	else:
		Banner.DirRunnerBanner()
		Banner.FingerprintModeBanner(URL)

		if not validators.url(URL):
			print(TerminalColor.LightRed +"Invalid url!"+TerminalColor.Reset)
			sys.exit()

		try:

			print(f'[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking connection for {URL}{TerminalColor.Reset}')
			headers={"User-Agent":f"{UserAgent}"}
			res = requests.get(URL,headers=headers,allow_redirects=False,timeout=5)
			print(f'[{TerminalColor.Green}+{TerminalColor.Reset}]{TerminalColor.Green} Connection OK!{TerminalColor.Reset}')

		except requests.exceptions.Timeout:
			print(f"{TerminalColor.Red}Timeout for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.ConnectionError:
			print(f"{TerminalColor.Red}Connection error for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.TooManyRedirects:
			print(f"{TerminalColor.Red}Too may redirect for {URL}{TerminalColor.Reset}")
			sys.exit()
		except requests.exceptions.RequestException as e:
			raise SystemExit(e)
			sys.exit()

		Modules.FingerprintRequest(URL,USER_AGENT)


def DnsModule(Domain="",Wordlist="",Thread=10):

	print(TerminalColor.Green +'dns attack mode selected'+TerminalColor.Reset)

	if not Domain:
		print(TerminalColor.LightRed +"some args are empty!"+TerminalColor.Reset)
		Banner.Help()
		sys.exit()
	else:
		if not Wordlist:
			Wordlist="wordlist/subdomains.txt"

		file = Path(Wordlist)
		if not file.is_file():
			print(TerminalColor.LightRed +f"file {Wordlist} not found!"+TerminalColor.Reset)
			sys.exit()

		Banner.DirRunnerBanner()
		Banner.DnsModeBanner(Domain,Thread,Wordlist)
		if not validators.domain(Domain):
			print(TerminalColor.LightRed +"Invalid domain!"+TerminalColor.Reset)
			sys.exit()

		Tasks.DnsModuleThreads(Domain,Wordlist,Thread)


def main():

	parser = argparse.ArgumentParser(prog='PROG')
	parser.add_argument('-u','--url',help='Set target URL on Dir enumeration attack')
	parser.add_argument('-a','--user-agent',default='DirRunner v1.0',help='Set user-agent')
	parser.add_argument('-d','--domain',help='Set target domain on DNS enumeration attack')
	parser.add_argument('-w','--wordlist',help='Set wordlist file')
	parser.add_argument('-s','--status-code',default='200,301,302,401,403,405',help='Set the status code to print (200,301)')
	parser.add_argument('-x','--exts',help='Set target extension files (php,txt,html)')
	parser.add_argument('-t','--thread',default='10',help='Set THREADS')
	parser.add_argument('-m','--method',default='GET',help='Set method (GET/POST), GET by default.')
	parser.add_argument('mode',help='Set attack mode (dir,dns,fingerprint,file)')
	#parser.add_argument('-r','--recursive-folder',default=False,help='Set recursive mode for new directories founds.')
	args = parser.parse_args()

	TARGET_URL=args.url
	WORDLIST_FILE=args.wordlist
	TARGET_DOMAIN=args.domain
	STATUS_CODE=args.status_code.split(",")
	EXTENSIONS=[]
	if args.exts:
		EXTENSIONS=args.exts.split(",")

	THREADS=args.thread
	METHOD=args.method.split(",")
	ATTACK_MODE=args.mode
	USER_AGENT=args.user_agent

	#recursive_folder_mode=args.recursive_folder

	ResponseCode.init(STATUS_CODE)
	
	if ATTACK_MODE == 'dir':
		DirectoryModule(TARGET_URL,METHOD,WORDLIST_FILE,STATUS_CODE,USER_AGENT,THREADS)
	if ATTACK_MODE == 'dns':
		DnsModule(TARGET_DOMAIN,WORDLIST_FILE,THREADS)
	if ATTACK_MODE == 'fingerprint':
		FingerprintModule(TARGET_URL)
	if ATTACK_MODE == 'file':
		FileModule(TARGET_URL,WORDLIST_FILE,STATUS_CODE,EXTENSIONS,USER_AGENT,THREADS)

if __name__=='__main__':
	main()




	