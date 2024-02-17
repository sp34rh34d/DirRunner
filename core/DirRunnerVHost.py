from core.main import *
from pathlib import Path
import concurrent.futures,urllib3,requests
from tabulate import tabulate
urllib3.disable_warnings()

class vhost_option :
	module_name="virtual host enumeration"
	wordlist="wordlist/subdomains.txt"
	tls_validation=True
	target_url=""
	target_domain=""
	threads=10
	cookie=""
	user_agent="DirRunner v1.2"
	username=""
	password=""
	founds=[]

class vhost_module:

	def menu():
		menu=[]
		module=[f'{c.Blue}#{c.Reset}','Module',f'{c.Green}{vhost_option.module_name}{c.Reset}']
		menu.append([f'{c.Red}1{c.Reset}','Target url',f'{c.Green}{vhost_option.target_url}{c.Reset} ({c.Orange}required{c.Reset})'])
		menu.append([f'{c.Red}2{c.Reset}','Target domain',f'{c.Green}{vhost_option.target_domain}{c.Reset} ({c.Orange}required{c.Reset})'])
		menu.append([f"{c.Red}3{c.Reset}", "Wordlist" ,f"{c.Green}{vhost_option.wordlist}{c.Reset}"])
		menu.append([f"{c.Red}4{c.Reset}", "User agent" ,f"{c.Green}{vhost_option.user_agent}{c.Reset}"])
		menu.append([f"{c.Red}5{c.Reset}", "Threads" ,f"{c.Green}{vhost_option.threads}{c.Reset}"])
		menu.append([f"{c.Red}6{c.Reset}", "TLS validation" ,f"{c.Green}{vhost_option.tls_validation}{c.Reset}"])
		menu.append([f"{c.Red}7{c.Reset}", "Cookies" ,f"{c.Green}{vhost_option.cookie}{c.Reset}"])
		menu.append([f"{c.Red}8{c.Reset}", "Username" ,f"{c.Green}{vhost_option.username}{c.Reset}"])
		menu.append([f"{c.Red}9{c.Reset}", "Password" ,f"{c.Green}{vhost_option.password}{c.Reset}"])
		menu.append([f'{c.Red}97{c.Reset}','Run',''])
		menu.append([f'{c.Red}98{c.Reset}','Export results',''])
		menu.append([f'{c.Red}99{c.Reset}','Back',''])
		print(tabulate(menu, module,  tablefmt="grid",numalign="left"))

	def main():
		while True:
			Banner.display_banner=True
			Banner.display()
			vhost_module.menu()
			prompt=''
			try:
				prompt=int(input("🌐 : "))
			except:
				pass

			if prompt==99:
				Banner.display_banner=True
				break
			elif prompt==1:
				vhost_option.target_url=input("enter target url: ").replace(" ","")
			elif prompt==2:
				vhost_option.target_domain=input("enter target domain: ").replace(" ","")
			elif prompt==3:
				vhost_option.wordlist=input("enter wordlist path: ")
			elif prompt==4:
				vhost_option.user_agent=input("enter user agent: ")
				if not vhost_option.user_agent:
					vhost_option.user_agent="DirRunner v1.2"
			elif prompt==5:
				try:
					vhost_option.threads=int(input("enter threads: "))
				except:
					vhost_option.threads=10
			elif prompt==6:
				validate=input("do you want to validate tls (y/n): ")
				vhost_option.tls_validation=True
				if validate=='n':
					vhost_option.tls_validation=False
			elif prompt==7:
				vhost_option.cookie=input("enter cookies: ")
			elif prompt==8:
				vhost_option.username=input("enter username: ")
			elif prompt==9:
				vhost_option.password=input("enter password: ")
			elif prompt==97:
				vhost_module.run()
			elif prompt==98:
				vhost_module.export()
			else:
				message.error("invalid input!")


	def run():
		run=True
		if not vhost_option.target_url or not validator.url(vhost_option.target_url):
			message.error("invalid target url!")
			run=False
			return
		if not vhost_option.target_domain or not validator.domain(vhost_option.target_domain):
			message.error("invalid target domain!")
			return

		if not vhost_option.wordlist:
			vhost_option.wordlist='wordlist/directory.txt'

		file = Path(vhost_option.wordlist)
		if not file.is_file():
			message.error(f"file {vhost_option.wordlist} not found!")
			run=False
			return

			
		if run:
			try:
				vhost_option.founds=[]
				message.info("running virtual host enumeration...")
				with concurrent.futures.ThreadPoolExecutor(max_workers=int(vhost_option.threads)) as executor:
					f = open(vhost_option.wordlist,'r')
					future_to_url = {executor.submit(vhost_module.send_http_request,word): word for word in f.read().split("\n")}
		
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
			message.error("virtual host enumeration not started!")

	def send_http_request(line):
		hostname=f"{line}.{vhost_option.target_domain}".replace("..",".")
		print(f'{hostname}                             ',end="\r")

		custom_headers={
			"User-Agent":vhost_option.user_agent,
			"cookie":vhost_option.cookie,
			"Host":hostname
		}

		auth = None
		if vhost_option.username and vhost_option.password:
			auth = (vhost_option.username, vhost_option.password)

		try:
			custom_request=requests.get(
				vhost_option.target_url,
				auth=auth,
				headers=custom_headers,
				timeout=2,
				verify=vhost_option.tls_validation
			)
			
			if custom_request.status_code==200:
				print(f"[{c.Green}200{c.Reset}] {hostname}")
				vhost_option.founds.append(hostname)
		except:
			pass

	def export():
		message.info(f"exporting results on file {vhost_option.target_domain}.txt...")
		with open(f'{vhost_option.target_domain}.txt','w') as f:
			for item in vhost_option.founds:
				f.write(f'{item[0]}\n')
		message.success("done!")