#!/usr/bin/env python3
import requests
import argparse
import sys
import concurrent.futures
import socket
import urllib3
from urllib3.exceptions import InsecureRequestWarning
from builtwith import *

target_url=""
wordlist_file=""
status_code=[]
extensions=[]
threads=10
target_domain=""
method=[]
attack_mode=""
user_agent=""

code_200=False
code_201=False
code_301=False
code_302=False
code_400=False
code_401=False
code_403=False
code_404=False
code_405=False
code_500=False
code_503=False

class color:
	black = '\033[30m'
	red = '\033[31m'
	green = '\033[32m'
	orange = '\033[33m'
	blue = '\033[34m'
	purple = '\033[35m'
	cyan = '\033[36m'
	lightgrey = '\033[37m'
	darkgrey = '\033[90m'
	lightred = '\033[91m'
	lightgreen = '\033[92m'
	yellow = '\033[93m'
	lightblue = '\033[94m'
	pink = '\033[95m'
	lightcyan = '\033[96m'
	reset = '\033[0m'

def read_response_code(_res,_line,_method):
	match int(_res.status_code):
		case 200:
			if code_200==True:
				print(f'[{color.green}200{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                        ')
		case 201:
			if code_201==True:
				print(f'[{color.green}201{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                         ')
		case 301:
			if code_301==True:
				print(f'[{color.yellow}301{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset} -> [{color.blue}{_res.headers["Location"]}{color.reset}]')
		case 302:
			if code_302==True:
				print(f'[{color.yellow}302{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset} -> [{color.blue}{_res.headers["Location"]}{color.reset}]')
		case 400:
			if code_400==True:
				print(f'[{color.blue}400{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                          ')
		case 401:
			if code_401==True:
				print(f'[{color.blue}401{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                          ')
		case 403:
			if code_403==True:
				print(f'[{color.blue}403{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                          ')
		case 404:
			if code_404==True:
				print(f'[{color.blue}404{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                          ')
		case 405:
			if code_405==True:
				print(f'[{color.blue}405{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                          ')
		case 500:
			if code_500==True:
				print(f'[{color.red}500{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                           ')
		case 503:
			if code_503==True:
				print(f'[{color.red}503{color.reset}][{color.yellow}{_method}{color.reset}] {color.green}{_line}{color.reset}                           ')

def url_request(_url,line):

	headers={"User-Agent":f"{user_agent}"}
	url_to_request=''
	urllib3.disable_warnings()
	requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

	read_url = _url[len(_url) - 1]
	if read_url =='/':
		url_to_request=f'{_url}{line}'
	else:
		url_to_request=f'{_url}/{line}'
		print(f'processing: {line}                          ',end="\r")

	for m in method:
		match m:
			case "GET":
				res = requests.get(url_to_request,headers=headers,allow_redirects=False,timeout=5)
				read_response_code(res,line,"GET")

			case "POST":
				res = requests.post(url_to_request,headers=headers,allow_redirects=False,timeout=5)
				read_response_code(res,line,"POST")

			case "PUT":
				res = requests.put(url_to_request,headers=headers,allow_redirects=False,timeout=5)
				read_response_code(res,line,"PUT")

			case "HEAD":
				res = requests.head(url_to_request,headers=headers,allow_redirects=False,timeout=5)
				read_response_code(res,line,"HEAD")

			case "DELETE":
				res = requests.delete(url_to_request,headers=headers,allow_redirects=False,timeout=5)
				read_response_code(res,line,"DELETE")

			case "OPTION":
				res = requests.option(url_to_request,headers=headers,allow_redirects=False,timeout=5)
				read_response_code(res,line,"OPTION")



def file_discovery(_url,line):
	headers={"User-Agent":f"{user_agent}"}
	url_to_request=''
	urllib3.disable_warnings()
	requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

	read_url = _url[len(_url) - 1]
	if read_url =='/':
		url_to_request=f'{_url}{line}'
	else:
		url_to_request=f'{_url}/{line}'
		print(f'processing: {line}                          ',end="\r")

	for extension in extensions:
		built_url_to_request=f'{url_to_request}.{extension.replace(".","")}'
		res = requests.get(built_url_to_request,headers=headers,allow_redirects=False,timeout=5)

		if res.status_code==200:
			print(f'[{color.green}+{color.reset}] {color.green}{built_url_to_request}{color.reset}                        ')


def ThreadsFileMode():
	try:

		with concurrent.futures.ThreadPoolExecutor(max_workers=int(threads)) as executor:
			f = open(wordlist_file,'r')
			future_to_url = {executor.submit(file_discovery,target_url,word): word for word in f.read().split("\n")}
			for future in concurrent.futures.as_completed(future_to_url):
				try:
					data = future.result()
				except err:
					return 0

	except:
		print(f'{color.red}Process terminated!{color.reset}')

def ThreadsDirMode():
	try:
		with concurrent.futures.ThreadPoolExecutor(max_workers=int(threads)) as executor:
			f = open(wordlist_file,'r')
			future_to_url = {executor.submit(url_request,target_url,word): word for word in f.read().split("\n")}
			for future in concurrent.futures.as_completed(future_to_url):
				try:
					data = future.result()
				except err:
					return 0

	except:
		print(f'{color.red}Process terminated!{color.reset}')

def hostname_request(_domain,line):
	try:
		_domain_to_request=f'{line}.{_domain}'
		_hostname=_domain_to_request.replace("..",".")
		print(f'processing: {_hostname}                          ',end="\r")
		res=socket.gethostbyname(_hostname)
		print(f'[{color.green}+{color.reset}] {color.green}{_hostname}{color.reset} - [{color.yellow}{res}{color.reset}]                                   ')
	except:
		return 0

def ThreadsDnsMode():
	try:
		with concurrent.futures.ThreadPoolExecutor(max_workers=int(threads)) as executor:
			f = open(wordlist_file,'r')
			future_to_url = {executor.submit(hostname_request,target_domain,word): word for word in f.read().split("\n")}
			for future in concurrent.futures.as_completed(future_to_url):
				try:
					data = future.result()
				except err:
					return 0
	except:
		print(f'{color.red}Process terminated!{color.reset}')

def fingerprint_request():
	headers={"User-Agent":f"{user_agent}"}
	urllib3.disable_warnings()
	requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)
	res = requests.post(target_url,headers=headers,allow_redirects=False,timeout=15,verify=False)
	tech=builtwith(target_url)

	try:
		print(f"Status: {color.green}{res.status_code}{color.reset}")
		print(f"Web server: {color.green}{res.headers['Server']}{color.reset}")
		print(f"Content-Length: {color.green}{res.headers['Content-Length']}{color.reset}")
	except:
		return 0
	
	for f in tech:
		print(f'{f} : {color.green}{tech[f]}{color.reset}')

def banner():
	print(
		f""" 
 ;                                                                                                    
 ED.                                                                                                  
 E#Wi                                     :       L.             L.                     ,;            
 E###G.       t   j.          j.          Ef      EW:        ,ft EW:        ,ft       f#i  j.         
 E#fD#W;      Ej  EW,         EW,         E#t     E##;       t#E E##;       t#E     .E#t   EW,        
 E#t t##L     E#, E##j        E##j        E#t     E###t      t#E E###t      t#E    i#W,    E##j       
 E#t  .E#K,   E#t E###D.      E###D.      E#t     E#fE#f     t#E E#fE#f     t#E   L#D.     E###D.     
 E#t    j##f  E#t E#jG#W;     E#jG#W;     E#t fi  E#t D#G    t#E E#t D#G    t#E :K#Wfff;   E#jG#W;    
 E#t    :E#K: E#t E#t t##f    E#t t##f    E#t L#j E#t  f#E.  t#E E#t  f#E.  t#E i##WLLLLt  E#t t##f   
 E#t   t##L   E#t E#t  :K#E:  E#t  :K#E:  E#t L#L E#t   t#K: t#E E#t   t#K: t#E  .E#L      E#t  :K#E: 
 E#t .D#W;    E#t E#KDDDD###i E#KDDDD###i E#tf#E: E#t    ;#W,t#E E#t    ;#W,t#E    f#E:    E#KDDDD###i
 E#tiW#G.     E#t E#f,t#Wi,,, E#f,t#Wi,,, E###f   E#t     :K#D#E E#t     :K#D#E     ,WW;   E#f,t#Wi,,,
 E#K##i       E#t E#t  ;#W:   E#t  ;#W:   E#K,    E#t      .E##E E#t      .E##E      .D#;  E#t  ;#W:  
 E##D.        E#t DWi   ,KK:  DWi   ,KK:  EL      ..         G#E ..         G#E        tt  DWi   ,KK: 
 E#t          ,;.                         :                   fE             fE                       
 L:                                                            ,              ,
Coded by:{color.red} Adonis Izaguirre {color.reset} Email:{color.red} adonis.izaguirre@kapa7.com {color.reset}
Welcome to DirRunner v1.0
======================================================================================================""")

def help():
	print(
		"""
-u: set target url
-d: set target domain
-a: set user-agent
-x: set target extensions files (php,txt,html)
-s: set the status code to print (200,301)
-w: set wordlist
-t: set threads
-m: set method (GET/POST), GET by default.
-h: show this message

dns enumeration
use: python3 DirRunner.py dns -d domain.com -w wordlist.txt

dir enumeration
use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt

print only status code 200 and 301
use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -s 200,301

file detection by extensions
use: python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x php,txt

fingerprint
user: python3 DirRunner.py fingerprint -u https://www.domain.com """
		)

parser = argparse.ArgumentParser(prog='PROG')
parser.add_argument('-u','--url',help='Set target URL on Dir enumeration attack')
parser.add_argument('-a','--user-agent',default='DirRunner v1.0',help='Set user-agent')
parser.add_argument('-d','--domain',help='Set target domain on DNS enumeration attack')
parser.add_argument('-w','--wordlist',help='Set wordlist file')
parser.add_argument('-s','--status-code',default='200,301,302,401,403,405',help='Set the status code to print (200,301)')
parser.add_argument('-x','--extensions-file',help='Set target extensions files (php,txt,html)')
parser.add_argument('-t','--thread',default='10',help='Set threads')
parser.add_argument('-m','--method',default='GET',help='Set method (GET/POST), GET by default.')
parser.add_argument('mode',help='Set attack mode (dir,dns,fingerprint,file)')
args = parser.parse_args()

target_url=args.url
wordlist_file=args.wordlist
target_domain=args.domain
status_code=args.status_code.split(",")

if args.extensions_file:
	extensions=args.extensions_file.split(",")

threads=args.thread
method=args.method.split(",")
attack_mode=args.mode
user_agent=args.user_agent

for s in status_code:
	match int(s):
		case 200:
			code_200=True
		case 201:
			code_201=True
		case 301:
			code_301=True
		case 302:
			code_302=True
		case 400:
			code_400=True
		case 401:
			code_401=True
		case 403:
			code_403=True
		case 404:
			code_404=True
		case 405:
			code_405=True
		case 500:
			code_500=True
		case 503:
			code_503=True


if attack_mode == 'dir':
	print(color.green + 'dir attack mode selected'+color.reset)
	if not target_url or not wordlist_file:
		print(color.lightred +"some args are empty!"+color.reset)
		help()
		sys.exit()
	else:
		banner()
		print(f"""Target: {color.green}{target_url}{color.reset}
Method: {color.green}{method}{color.reset}
Attack mode: {color.green}{attack_mode}{color.reset}
Threads: {color.green}{threads}{color.reset}
Status code: {color.green}{status_code}{color.reset}
User-agent: {color.green}{user_agent}{color.reset}
Extension: {color.green}{extensions}{color.reset}
Wordlist file: {color.green}{wordlist_file}{color.reset}
======================================================================================================""")
		ThreadsDirMode()

if attack_mode == 'dns':
	print(color.green +'dns attack mode selected'+color.reset)
	if not target_domain or not wordlist_file:
		print(color.lightred +"some args are empty!"+color.reset)
		help()
		sys.exit()
	else:
		banner()
		print(f"""Target: {color.green}{target_domain}{color.reset}
Attack mode: {color.green}{attack_mode}{color.reset}
Threads: {color.green}{threads}{color.reset}
Wordlist file: {color.green}{wordlist_file}{color.reset}
======================================================================================================""")
		ThreadsDnsMode()

if attack_mode == 'fingerprint':
	print(color.green +'fingerprint mode selected'+color.reset)
	if not target_url:
		print(color.lightred +"some args are empty!"+color.reset)
		help()
		sys.exit()
	else:
		banner()
		print(f"""Target: {color.green}{target_url}{color.reset}
Attack mode: {color.green}{attack_mode}{color.reset}
======================================================================================================""")
		fingerprint_request()

if attack_mode == 'file':
	print(color.green +'file discovery mode selected'+color.reset)
	if not target_url or not wordlist_file or not extensions:
		print(color.lightred +"some args are empty!"+color.reset)
		help()
		sys.exit()
	else:
		banner()
		print(f"""Target: {color.green}{target_url}{color.reset}
Method: {color.green}{method}{color.reset}
Attack mode: {color.green}{attack_mode}{color.reset}
Threads: {color.green}{threads}{color.reset}
User-agent: {color.green}{user_agent}{color.reset}
Extension: {color.green}{extensions}{color.reset}
Wordlist file: {color.green}{wordlist_file}{color.reset}
======================================================================================================""")
		ThreadsFileMode()
