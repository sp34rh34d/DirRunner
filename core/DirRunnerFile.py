from core.main import *
from pathlib import Path
import string,random,requests,urllib3,concurrent.futures,readline
urllib3.disable_warnings()
from tabulate import tabulate

class file_option:
	module_name="File enumeration"
	target_url=""
	wordlist="wordlist/filename.txt"
	user_agent="DirRunner v1.2"
	threads=10
	extensions=['txt','php']
	tls_validation=True
	timeout=10
	cookie=""
	username=""
	password=""
	founds=[]

class file_module:

	def menu():
		menu=[]
		module=[f'{c.Blue}#{c.Reset}','Module',f'{c.Green}{file_option.module_name}{c.Reset}']
		menu.append([f'{c.Red}1{c.Reset}','Target url',f'{c.Green}{file_option.target_url}{c.Reset} ({c.Orange}required{c.Reset})'])
		menu.append([f"{c.Red}2{c.Reset}", "Wordlist" ,f"{c.Green}{file_option.wordlist}{c.Reset}"])
		menu.append([f"{c.Red}3{c.Reset}", "User agent" ,f"{c.Green}{file_option.user_agent}{c.Reset}"])
		menu.append([f"{c.Red}4{c.Reset}", "Threads" ,f"{c.Green}{file_option.threads}{c.Reset}"])
		menu.append([f"{c.Red}5{c.Reset}", "Extensions" ,f"{c.Green}{file_option.extensions}{c.Reset}"])
		menu.append([f"{c.Red}6{c.Reset}", "TLS validation" ,f"{c.Green}{file_option.tls_validation}{c.Reset}"])
		menu.append([f"{c.Red}7{c.Reset}", "TimeOut" ,f"{c.Green}{file_option.timeout}{c.Reset}"])
		menu.append([f"{c.Red}8{c.Reset}", "Cookies" ,f"{c.Green}{file_option.cookie}{c.Reset}"])
		menu.append([f"{c.Red}9{c.Reset}", "Username" ,f"{c.Green}{file_option.username}{c.Reset}"])
		menu.append([f"{c.Red}10{c.Reset}", "Password" ,f"{c.Green}{file_option.password}{c.Reset}"])
		menu.append([f'{c.Red}98{c.Reset}','Run',''])
		menu.append([f'{c.Red}97{c.Reset}','Export results',''])
		menu.append([f'{c.Red}99{c.Reset}','Back',''])
		print(tabulate(menu, module,  tablefmt="grid",numalign="left"))

	def main():
		while True:
			Banner.display_banner=True
			Banner.display()
			file_module.menu()
			prompt=''
			try:
				prompt=int(input("🤐: "))
			except:
				pass

			if prompt==99:
				Banner.display_banner=True
				break
			elif prompt==1:
				url=input("enter target url: ").replace(" ","")
				if url.endswith("/"):
					file_option.target_url=url
				else:
					file_option.target_url=url+"/"
			elif prompt==2:
				file_option.wordlist=input("enter wordlist path: ")
			elif prompt==3:
				file_option.user_agent=input("enter user agent: ")
				if not file_option.user_agent:
					file_option.user_agent='DirRunner v1.2'
			elif prompt==4:
				try:
					file_option.threads=int(input("enter threads: "))
				except:
					file_option.threads=10
			elif prompt==5:
				file_option.extensions=input("enter extension file (txt,php): ").split(',')
			elif prompt==6:
				validate=input("do you want to validate tls (y/n): ")
				file_option.tls_validation=True
				if validate=='n':
					file_option.tls_validation=False
			elif prompt==7:
				try:
					file_option.timeout=int(input("enter timeout (seconds): "))
				except:
					file_option.timeout=10
			elif prompt==8:
				file_option.cookie=input("enter cookies: ")
			elif prompt==9:
				file_option.username=input("enter username: ")
			elif prompt==10:
				file_option.password=input("enter password: ")
			elif prompt==97:
				file_module.run()
			elif prompt==98:
				file_module.export()
			else:
				message.error("invalid option!")


	def run():
		run=True
		if not validator.url(file_option.target_url):
			message.error("invalid target url!")
			run=False
			return

		if not file_option.wordlist:
			file_option.wordlist='wordlist/filename.txt'

		file = Path(file_option.wordlist)
		if not file.is_file():
			message.error(f"file {file_option.wordlist} not found!")
			run=False
			return
		
		if not file_module.validate_connection():
			message.error("connection error detected!")
			run=False
			return
		
		if run:
			try:
				file_option.founds=[]
				message.info("running file enumeration...")
				with concurrent.futures.ThreadPoolExecutor(max_workers=int(file_option.threads)) as executor:
					f = open(file_option.wordlist,'r')
					future_to_url = {executor.submit(file_module.request,word): word for word in f.read().split("\n")}
		
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
			message.error("file enumeration not started!")


	def validate_connection():
		try:
			message.info(f"checking connection for target {file_option.target_url}")
			custom_headers={
				"User-Agent":file_option.user_agent,
				"cookie":file_option.cookie,
			}

			auth = None
			if file_option.username and file_option.password:
				auth = (file_option.username, file_option.password)

			custom_request = requests.get(
				file_option.target_url,
				auth=auth,
				headers=custom_headers,
				timeout=file_option.timeout,
				verify=file_option.tls_validation
				)
			message.info("done!")
			return True
		except requests.exceptions.Timeout:
			message.error(f"timeout for {file_option.target_url}")
			return False
		except requests.exceptions.SSLError:
			message.error("ssl verification error!")
			return False
		except requests.exceptions.TooManyRedirects:
			message.error(f"too may redirect for {file_option.target_url}")
			return False
		except requests.exceptions.ConnectionError as e:
			message.error(f"connection error: {e}")
			return False
		except requests.exceptions.RequestException as e:
			message.error(e)
			return False
		
		
	def RandomStrings(size=28, chars=string.ascii_lowercase + string.digits):
		return ''.join(random.choice(chars) for _ in range(size))
	
	def request(line):
		for ext in file_option.extensions:
			filename=f"{line}.{ext}".replace("..",".")
			url=f"{file_option.target_url}{filename}".replace("..",".")
			print(f'{filename}                             ',end="\r")
			file_module.send_http_request(url)
				
		
	def send_http_request(url):
		custom_headers={
			"User-Agent":file_option.user_agent,
			"cookie":file_option.cookie,
		}

		auth = None
		if file_option.username and file_option.password:
			auth = (file_option.username, file_option.password)

		try:
			custom_request=requests.get(
				url,
				auth=auth,
				headers=custom_headers,
				timeout=file_option.timeout,
				verify=file_option.tls_validation
			)
			
			if custom_request.status_code==200:
				print(f"[{c.Green}200{c.Reset}] {url} [size: {custom_request.headers['Content-Length']}]")
				file_option.founds.append([url,custom_request.headers['Content-Length']])
		except:
			pass

	def export():
		random_name=file_module.RandomStrings()
		message.info(f"exporting results on file {random_name}.txt...")
		with open(f'{random_name}.txt','w') as f:
			for item in file_option.founds:
				f.write(f'{item[0]} {item[1]}\n')
		message.success("done!")