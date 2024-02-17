from core.main import *
import requests,urllib3
from builtwith import *
urllib3.disable_warnings()
from tabulate import tabulate

class fingerprint_option:
	module_name="Fingerprint module" 
	target_url=""
	user_agent="DirRunner v1.2"
	tls_validation=True
	timeout=15
	cookie=""
	username=""
	password=""

class fingerprint_module:
	def menu():
		menu=[]
		module=[f'{c.Blue}#{c.Reset}','Module',f'{c.Green}{fingerprint_option.module_name}{c.Reset}']
		menu.append([f'{c.Red}1{c.Reset}','Target url',f'{c.Green}{fingerprint_option.target_url}{c.Reset} ({c.Orange}required{c.Reset})'])
		menu.append([f"{c.Red}2{c.Reset}", "User agent" ,f"{c.Green}{fingerprint_option.user_agent}{c.Reset}"])
		menu.append([f"{c.Red}3{c.Reset}", "TLS validation" ,f"{c.Green}{fingerprint_option.tls_validation}{c.Reset}"])
		menu.append([f"{c.Red}4{c.Reset}", "TimeOut" ,f"{c.Green}{fingerprint_option.timeout}{c.Reset}"])
		menu.append([f"{c.Red}5{c.Reset}", "Cookies" ,f"{c.Green}{fingerprint_option.cookie}{c.Reset}"])
		menu.append([f"{c.Red}6{c.Reset}", "Username" ,f"{c.Green}{fingerprint_option.username}{c.Reset}"])
		menu.append([f"{c.Red}7{c.Reset}", "Password" ,f"{c.Green}{fingerprint_option.password}{c.Reset}"])
		menu.append([f'{c.Red}97{c.Reset}','Run',''])
		menu.append([f'{c.Red}99{c.Reset}','Back',''])
		print(tabulate(menu, module,  tablefmt="grid",numalign="left"))

	def main():
		while True:
			Banner.display_banner=True
			Banner.display()
			fingerprint_module.menu()
			prompt=''
			try:
				prompt=int(input("🐾: "))
			except:
				pass

			if prompt==99:
				Banner.display_banner=True
				break
			elif prompt==1:
				fingerprint_option.target_url=input("enter target url: ").replace(" ","")
			elif prompt==2:
				fingerprint_option.user_agent=input("enter user agent: ")
				if not fingerprint_option.user_agent:
					fingerprint_option.user_agent='DirRunner v1.2'
			elif prompt==3:
				validate=input("do you want to validate tls (y/n): ")
				fingerprint_option.tls_validation=True
				if validate=='n':
					fingerprint_option.tls_validation=False
			elif prompt==4:
				try:
					fingerprint_option.timeout=int(input("enter timeout (seconds): "))
				except:
					fingerprint_option.timeout=10
			elif prompt==5:
				fingerprint_option.cookie=input("enter cookies: ")
			elif prompt==6:
				fingerprint_option.username=input("enter username: ")
			elif prompt==7:
				fingerprint_option.password=input("enter password: ")
			elif prompt==97:
				fingerprint_module.run()
			else:
				message.error("invalid option!")

	def run():
		run=True
		if not validator.url(fingerprint_option.target_url):
			message.error("invalid target url!")
			run=False
			return
		
		if not fingerprint_module.validate_connection():
			message.error("connection error detected!")
			run=False
			return		
		
		if run:
			message.info("starting fingerprint module")
			fingerprint_module.send_http_request()
		else:
			message.error("fingerprint module not started!")

	def validate_connection():
		try:
			message.info(f"checking connection for target {fingerprint_option.target_url}")
			custom_headers={
				"User-Agent":fingerprint_option.user_agent,
				"cookie":fingerprint_option.cookie,
			}

			auth = None
			if fingerprint_option.username and fingerprint_option.password:
				auth = (fingerprint_option.username, fingerprint_option.password)

			custom_request = requests.get(
				fingerprint_option.target_url,
				auth=auth,
				headers=custom_headers,
				timeout=fingerprint_option.timeout,
				verify=fingerprint_option.tls_validation
				)
			message.info("done!")
			return True
		except requests.exceptions.Timeout:
			message.error(f"timeout for {fingerprint_option.target_url}")
			return False
		except requests.exceptions.SSLError:
			message.error("ssl verification error!")
			return False
		except requests.exceptions.TooManyRedirects:
			message.error(f"too may redirect for {fingerprint_option.target_url}")
			return False
		except requests.exceptions.ConnectionError as e:
			message.error(f"connection error: {e}")
			return False
		except requests.exceptions.RequestException as e:
			message.error(e)
			return False
	
	def send_http_request():
		try:
			custom_headers={
				"User-Agent":fingerprint_option.user_agent,
				"cookie":fingerprint_option.cookie,
			}

			auth = None
			if fingerprint_option.username and fingerprint_option.password:
				auth = (fingerprint_option.username, fingerprint_option.password)

			custom_request = requests.get(
				fingerprint_option.target_url,
				auth=auth,
				headers=custom_headers,
				timeout=fingerprint_option.timeout,
				verify=fingerprint_option.tls_validation
			)

		except requests.exceptions.Timeout:
			message.error(f"timeout for {fingerprint_option.target_url}")
			return
		except requests.exceptions.ConnectionError:
			message.error(f"connection error for {fingerprint_option.target_url}")
			return
		except requests.exceptions.TooManyRedirects:
			message.error(f"too may redirect for {fingerprint_option.target_url}")
			return
		except requests.exceptions.SSLError:
			message.error(f"ssl verification error for {fingerprint_option.target_url}")
			return

		print(f"Status: {c.Green}{custom_request.status_code}{c.Reset}")
		for header in custom_request.headers:
			print(f"{header}: {c.Green}{custom_request.headers[header]}{c.Reset}")

	
		try:
			print(f"{c.Orange}detecting framework...{c.Reset}")
			tech=builtwith(fingerprint_option.target_url)
			for f in tech:
				print(f'{f} : {c.Green}{tech[f]}{c.Reset}')

			print(f"{c.Green}Done!{c.Reset}")
		except:
			print(f'{c.Red}Error trying to get web tech!{c.Reset}')








