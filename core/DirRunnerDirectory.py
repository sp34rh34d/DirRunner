from core.main import *
import requests
from pathlib import Path
import concurrent.futures
import string
import random
import urllib3
from tabulate import tabulate
urllib3.disable_warnings()

class directory_option :
	module_name="Directory enumeration"
	target_url=""
	method='get'
	wordlist="wordlist/directory.txt"
	user_agent="DirRunner v1.2"
	threads=10
	status_code=[200,301,302]
	tls_validation=True
	follow_redirects=False
	timeout=15
	cookie=""
	username=""
	password=""
	valid_http_methods=['get','post','head','delete','put']
	founds=[]

class directory_module:

	def menu():
		menu=[]
		module=[f'{c.Blue}#{c.Reset}','Module',f'{c.Green}{directory_option.module_name}{c.Reset}']
		menu.append([f'{c.Red}1{c.Reset}','Target',f'{c.Green}{directory_option.target_url}{c.Reset} ({c.Orange}required{c.Reset})'])
		menu.append([f'{c.Red}2{c.Reset}','Method',f'{c.Green}{directory_option.method}{c.Reset}'])
		menu.append([f"{c.Red}3{c.Reset}", "Wordlist" ,f"{c.Green}{directory_option.wordlist}{c.Reset}"])
		menu.append([f"{c.Red}4{c.Reset}", "User agent" ,f"{c.Green}{directory_option.user_agent}{c.Reset}"])
		menu.append([f"{c.Red}5{c.Reset}", "Threads" ,f"{c.Green}{directory_option.threads}{c.Reset}"])
		menu.append([f"{c.Red}6{c.Reset}", "Show status code" ,f"{c.Green}{directory_option.status_code}{c.Reset}"])
		menu.append([f"{c.Red}7{c.Reset}", "TLS validation" ,f"{c.Green}{directory_option.tls_validation}{c.Reset}"])
		menu.append([f"{c.Red}8{c.Reset}", "Follow redirect" ,f"{c.Green}{directory_option.follow_redirects}{c.Reset}"])
		menu.append([f"{c.Red}9{c.Reset}", "TimeOut" ,f"{c.Green}{directory_option.timeout}{c.Reset}"])
		menu.append([f"{c.Red}10{c.Reset}", "Cookies" ,f"{c.Green}{directory_option.cookie}{c.Reset}"])
		menu.append([f"{c.Red}11{c.Reset}", "Username" ,f"{c.Green}{directory_option.username}{c.Reset}"])
		menu.append([f"{c.Red}12{c.Reset}", "Password" ,f"{c.Green}{directory_option.password}{c.Reset}"])
		menu.append([f'{c.Red}98{c.Reset}','Run',''])
		menu.append([f'{c.Red}97{c.Reset}','Export results',''])
		menu.append([f'{c.Red}99{c.Reset}','Back',''])
		print(tabulate(menu, module,  tablefmt="grid",numalign="left"))

	def main():
		while True:
			Banner.display_banner=True
			Banner.display()
			directory_module.menu()
			prompt=''
			try:
				prompt=int(input("🗂️ : "))
			except:
				pass

			if prompt==99:
				Banner.display_banner=True
				break
			elif prompt==1:
				url=input("enter target url: ").replace(" ","")
				if url.endswith("/"):
					directory_option.target_url=url
				else:
					directory_option.target_url=url+"/"
			elif prompt==2:
				try:
					method=input("enter http method: ")
					directory_option.method=method
					if not method:
						directory_option.methods='get'
				except:
					directory_option.method='get'
			elif prompt==3:
				directory_option.wordlist=input("enter wordlist path: ")
			elif prompt==4:
				directory_option.user_agent=input("enter user agent: ")
			elif prompt==5:
				try:
					directory_option.threads=int(input("enter threads: "))
				except:
					directory_option.threads=10
			elif prompt==6:
				try:
					directory_option.status_code=input("enter status code to display (200,301): ").split(',')
				except:
					directory_option.status_code=[200,301,302]
			elif prompt==7:
				validate=input("do you want to validate tls (y/n): ")
				directory_option.tls_validation=True
				if validate=='n':
					directory_option.tls_validation=False
			elif prompt==8:
				validate=input("follow redirects (y/n): ")
				directory_option.follow_redirects=True
				if validate=='n':
					directory_option.follow_redirects=False

			elif prompt==9:
				try:
					directory_option.timeout=int(input("enter timeout (seconds): "))
				except:
					directory_option.timeout=15
			elif prompt==10:
				directory_option.cookie=input("enter cookies: ")
			elif prompt==11:
				directory_option.username=input("enter username: ")
			elif prompt==12:
				directory_option.password=input("enter password: ")
			elif prompt==97:
				directory_module.run()
			elif prompt==98:
				directory_module.export()
			else:
				message.error("invalid input!")


	def run():
		message.info("starting directory enumeration...")
		run=True
		if not directory_option.target_url or not validator.url(directory_option.target_url):
			message.error("invalid target url!")
			run=False
			return
		
		if not directory_option.method.lower() in directory_option.valid_http_methods:
			message.error("invalid http method detected!")
			run=False
			return
		
		for code in directory_option.status_code:
			if not int(code) >199 or not int(code) <600: 
				message.error("invalid status code detected!")
				run=False
				return
		if not directory_option.wordlist:
			directory_option.wordlist='wordlist/directory.txt'

		file = Path(directory_option.wordlist)
		if not file.is_file():
			message.error(f"file {directory_option.wordlist} not found!")
			run=False
			return

		if not directory_module.validate_connection():
			message.error("connection error detected!")
			return

		for code in directory_option.status_code:
			if directory_module.check_non_existing_url(int(code)):
				run=False
				return
			message.info("done!")
			
		if run:
			try:
				directory_option.founds=[]
				message.info("running directory enumeration...")
				with concurrent.futures.ThreadPoolExecutor(max_workers=int(directory_option.threads)) as executor:
					f = open(directory_option.wordlist,'r')
					future_to_url = {executor.submit(directory_module.request,word): word for word in f.read().split("\n")}
		
					for future in concurrent.futures.as_completed(future_to_url):
						future.result()
			except KeyboardInterrupt:
				message.error("Process terminated, Ctrl C!")
				for future in future_to_url:
					future.cancel()
			finally:
				message.info("wait...")
				for future in future_to_url:
					future.cancel()
				message.info("done!")
		else:
			message.error("directory enumeration not started!")


	def validate_connection():
		try:
			message.info(f"checking connection for target {directory_option.target_url}")
			custom_headers={
				"User-Agent":directory_option.user_agent,
				"cookie":directory_option.cookie,
			}

			auth = None
			if directory_option.username and directory_option.password:
				auth = (directory_option.username, directory_option.password)

			custom_request = requests.get(
				directory_option.target_url,
				auth=auth,
				headers=custom_headers,
				allow_redirects=directory_option.follow_redirects,
				timeout=directory_option.timeout,
				verify=directory_option.tls_validation
				)
			message.info("done!")
			return True
		except requests.exceptions.Timeout:
			message.error(f"timeout for {directory_option.target_url}")
			return False
		except requests.exceptions.SSLError:
			message.error("ssl verification error!")
			return False
		except requests.exceptions.TooManyRedirects:
			message.error(f"too may redirect for {directory_option.target_url}")
			return False
		except requests.exceptions.ConnectionError as e:
			message.error(f"connection error: {e}")
			return False
		except requests.exceptions.RequestException as e:
			message.error(e)
			return False

	def RandomStrings(size=28, chars=string.ascii_lowercase + string.digits):
		return ''.join(random.choice(chars) for _ in range(size))
		
	def check_non_existing_url(code):
		random_directory=directory_module.RandomStrings()
		message.info(f"checking for non existing directory {directory_option.target_url}{random_directory} for status code {code}")
		url=f"{directory_option.target_url}{random_directory}"

		status_code,headers=directory_module.send_http_request(url)
		if status_code is not None:
			if status_code==code:
				message.warning(f"we got a valid http {directory_option.method} response from non existing directory, please exclude {code}")
				return True
			else:
				return False

	def send_http_request(url):
		custom_headers={
			"User-Agent":directory_option.user_agent,
			"cookie":directory_option.cookie,
		}

		auth = None
		if directory_option.username and directory_option.password:
			auth = (directory_option.username, directory_option.password)

		try:
				if directory_option.method.lower()=="get":
					custom_request=requests.get(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=directory_option.follow_redirects,
						timeout=directory_option.timeout,
						verify=directory_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers
				
				if directory_option.method.lower()=="post":
					custom_request=requests.post(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=directory_option.follow_redirects,
						timeout=directory_option.timeout,
						verify=directory_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers
				
				if directory_option.method.lower()=="head":
					custom_request=requests.head(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=directory_option.follow_redirects,
						timeout=directory_option.timeout,
						verify=directory_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers
				
				if directory_option.method.lower()=="delete":
					custom_request=requests.put(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=directory_option.follow_redirects,
						timeout=directory_option.timeout,
						verify=directory_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers
				
				if directory_option.method.lower()=="put":
					custom_request=requests.put(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=directory_option.follow_redirects,
						timeout=directory_option.timeout,
						verify=directory_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers

		except requests.exceptions.Timeout:
			message.error(f"connection timeout for {url}")
			return None, None
		except:
			return None, None

	def request(line):
		print(f'{line}                             ',end="\r")
		url=f"{directory_option.target_url}{line}/"
		status_code,headers=directory_module.send_http_request(url)
		if status_code is not None and status_code in directory_option.status_code:
			directory_module.output(status_code,headers,line)

	def output(code,header,line):
		if code >199 and code <300:
			print(f"[{c.Green}{code}{c.Reset}] /{line} [size: {header['Content-Length']}]")
		elif code >299 and code <400:
			print(f"[{c.Orange}{code}{c.Reset}] /{line} [size: {header['Content-Length']}] -> [{header['Location']}]")
		elif code >399 and code <500:
			print(f"[{c.Blue}{code}{c.Reset}] /{line} [size: {header['Content-Length']}]")
		elif code >499 and code <600:
			print(f"[{c.Red}{code}{c.Reset}] /{line} [size: {header['Content-Length']}]")

		try:
			item=[code,directory_option.method,f"{directory_option.target_url}{line}",header["Content-Length"],header["Location"]]
		except:
			item=[code,directory_option.method,f"{directory_option.target_url}{line}",header["Content-Length"],'']
		directory_option.founds.append(item)

	def export():
		random_name=directory_module.RandomStrings()
		message.info(f"exporting results on file {random_name}.txt...")
		with open(f'{random_name}.txt','w') as f:
			for item in directory_option.founds:
				f.write(f'{item[0]} {item[1]} {item[2]} {item[3]} {item[4]}\n')
		message.success("done!")



