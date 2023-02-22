import concurrent.futures
import socket
import requests
import sys
from builtwith import *
from core.dirRunnerMessage import *
import string
import random

class Tasks:
	def DnsModuleThreads(Domain="",Wordlist="",Threads=10):
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(Threads)) as executor:
				f = open(Wordlist,'r')
				future_to_url = {executor.submit(Modules.HostnameRequest,Domain,word): word for word in f.read().split("\n")}
		
				for future in concurrent.futures.as_completed(future_to_url):
					future.result()

		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}')
			sys.exit()

	def DirModuleThreads(URL="",Method=[],Wordlist="",UserAgent="",Threads=10):
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(Threads)) as executor:
				f = open(Wordlist,'r')
				future_to_url = {executor.submit(Modules.UrlRequest,URL,Method,UserAgent,word): word for word in f.read().split("\n")}

				for future in concurrent.futures.as_completed(future_to_url):
					future.result()

		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}')
			sys.exit()

	def FileModuleThreads(URL="",Wordlist="",UserAgent="",Exts=[],Threads=10):
		try:
			with concurrent.futures.ThreadPoolExecutor(max_workers=int(Threads)) as executor:
				f = open(Wordlist,'r')
				future_to_url = {executor.submit(Modules.FileDiscovery,URL,UserAgent,Exts,word): word for word in f.read().split("\n")}

				for future in concurrent.futures.as_completed(future_to_url):
					future.result()
		except KeyboardInterrupt:
			print(f'{TerminalColor.Red}Process terminated, Ctrl C!{TerminalColor.Reset}')
			sys.exit()


class Modules:

	def RandomStrings(size=32, chars=string.ascii_uppercase + string.digits):
		return ''.join(random.choice(chars) for _ in range(size))

	def NonExistingUrlCheck(URL,Method,UserAgent):

		HEADERS={"User-Agent":f"{UserAgent}"}

		RANDOM_URL=Modules.RandomStrings()
		READ_URL = URL[len(URL) - 1]
		res=requests

		if READ_URL =='/':
			URL_TO_REQUEST=f'{URL}{RANDOM_URL}'
		else:
			URL_TO_REQUEST=f'{URL}/{RANDOM_URL}'

		print(f"[{TerminalColor.Blue}!{TerminalColor.Reset}] {TerminalColor.Orange}Checking for non existing urls response!{TerminalColor.Reset}",end="\r")
		print(f"[{TerminalColor.Blue}!{TerminalColor.Reset}] Testing random url {TerminalColor.Orange}{URL_TO_REQUEST}{TerminalColor.Reset} ...",end="\r")

		for m in Method:
			try:
				if m == "GET":
					res = requests.get(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

				if m == "POST":
					res = requests.post(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

				if m == "PUT":
					res = requests.put(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

				if m == "HEAD":
					res = requests.head(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

				if m == "DELETE":
					res = requests.delete(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)

				if m == "OPTION":
					res = requests.option(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)


				if res.status_code==302:
					print(f"[{TerminalColor.Red}-{TerminalColor.Reset}] The website return a status code {TerminalColor.Orange}302{TerminalColor.Reset} for non existing urls {TerminalColor.Orange}{URL_TO_REQUEST}{TerminalColor.Reset}, please exclude this code from outputs. You can add following args '-s 200,301'")
					return True

			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Connection timeout for {URL_TO_REQUEST}{TerminalColor.Reset}")
			except:
				return 0

	def HostnameRequest(Domain="",Line=""):
		try:
			DOMAIN_TO_REQUEST=f'{Line}.{Domain}'
			HOSTNAME=DOMAIN_TO_REQUEST.replace("..",".")
			print(f'processing: {HOSTNAME}                          ',end="\r")
			res=socket.gethostbyname(HOSTNAME)
			print(f'[{TerminalColor.Green}+{TerminalColor.Reset}] {TerminalColor.Green}{HOSTNAME}{TerminalColor.Reset} - [{TerminalColor.Yellow}{res}{TerminalColor.Reset}]                                   ')
		except:
			return 0

	def FileDiscovery(URL="",UserAgent="",Exts=[],Line=""):
		HEADERS={"User-Agent":f"{UserAgent}"}

		READ_URL = URL[len(URL) - 1]
		if READ_URL =='/':
			URL_TO_REQUEST=f'{URL}{Line}'
		else:
			URL_TO_REQUEST=f'{URL}/{Line}'
		
		print(f'processing: {Line}                                    ',end="\r")

		for extension in Exts:
			BUILD_URL_TO_REQUEST=f'{URL_TO_REQUEST}.{extension.replace(".","")}'
			try:
				res = requests.get(BUILD_URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=15)

				if res.status_code==200:
					print(f'[{TerminalColor.Green}+{TerminalColor.Reset}] {TerminalColor.Green}{BUILD_URL_TO_REQUEST}{TerminalColor.Reset}                        ')
			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Timeout for {BUILD_URL_TO_REQUEST}{TerminalColor.Reset}")
			except:
				return 0


	def UrlRequest(URL="",Method=[],UserAgent="",Line=""):

		HEADERS={"User-Agent":f"{UserAgent}"}

		READ_URL = URL[len(URL) - 1]
		if READ_URL =='/':
			URL_TO_REQUEST=f'{URL}{Line}'
		else:
			URL_TO_REQUEST=f'{URL}/{Line}'
			print(f'processing: {Line}                                                                 ',end="\r")

		for m in Method:
			try:
				if m == "GET":
					res = requests.get(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
					ResponseCode.ReadResponseCode(res,Line,"GET")

				if m == "POST":
					res = requests.post(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
					ResponseCode.ReadResponseCode(res,Line,"POST")

				if m == "PUT":
					res = requests.put(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
					ResponseCode.ReadResponseCode(res,Line,"PUT")

				if m == "HEAD":
					res = requests.head(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
					ResponseCode.ReadResponseCode(res,Line,"HEAD")

				if m == "DELETE":
					res = requests.delete(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
					ResponseCode.ReadResponseCode(res,Line,"DELETE")

				if m == "OPTION":
					res = requests.option(URL_TO_REQUEST,headers=HEADERS,allow_redirects=False,timeout=10)
					ResponseCode.ReadResponseCode(res,Line,"OPTION")
	
			except requests.exceptions.Timeout:
				print(f"{TerminalColor.Red}Connection timeout for {URL_TO_REQUEST}{TerminalColor.Reset}")
			except:
				return 0


	def FingerprintRequest(URL="",UserAgent=""):
		
		headers={"User-Agent":f"{UserAgent}"}
		
		try:
			res = requests.get(URL,headers=headers,timeout=15)
		except requests.exceptions.Timeout:
	 		print(f"{TerminalColor.Red}Timeout for {URL}{TerminalColor.Reset}")
		except requests.exceptions.ConnectionError:
			print(f"{TerminalColor.Red}Connection error for {URL}{TerminalColor.Reset}")
		except requests.exceptions.TooManyRedirects:
	 		print(f"{TerminalColor.Red}Too may redirect for {URL}{TerminalColor.Reset}")

		try:
			print(f"Status: {TerminalColor.Green}{res.status_code}{TerminalColor.Reset}")
			print(f"Web server: {TerminalColor.Green}{res.headers['Server']}{TerminalColor.Reset}")
			print(f"Content-Length: {TerminalColor.Green}{res.headers['Content-Length']}{TerminalColor.Reset}")
		except:
			print("",end="\r")
	
		try:
			tech=builtwith(URL)
			for f in tech:
				print(f'{f} : {TerminalColor.Green}{tech[f]}{TerminalColor.Reset}')
		except:
			print(f'{TerminalColor.Red}Error trying to get web tech!{TerminalColor.Reset}')

