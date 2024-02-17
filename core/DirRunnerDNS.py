from core.main import *
from pathlib import Path
import readline, socket,concurrent.futures
from tabulate import tabulate

class dns_options:
	module_name="DNS Enumeration"
	wordlist="wordlist/subdomains.txt"
	target_domain=""
	threads=10
	founds=[]

class dns_module:

	def menu():
		menu=[]
		module=[f'{c.Blue}#{c.Reset}','Module',f'{c.Green}{dns_options.module_name}{c.Reset}']
		menu.append([f'{c.Red}1{c.Reset}','Target',f'{c.Green}{dns_options.target_domain}{c.Reset} ({c.Orange}required{c.Reset})'])
		menu.append([f'{c.Red}2{c.Reset}','Wordlist',f'{c.Green}{dns_options.wordlist}{c.Reset}'])
		menu.append([f"{c.Red}3{c.Reset}", "Threads" ,f"{c.Green}{dns_options.threads}{c.Reset}"])
		menu.append([f'{c.Red}97{c.Reset}','Run',''])
		menu.append([f'{c.Red}98{c.Reset}','Export results',''])
		menu.append([f'{c.Red}99{c.Reset}','Back',''])
		print(tabulate(menu, module,  tablefmt="grid",numalign="left",floatfmt=".2f"))
		
	def main():
		while True:
			Banner.display_banner=True
			Banner.display()
			dns_module.menu()
			prompt=''
			try:
				prompt=int(input("🌎: "))
			except:
				pass

			if prompt==1:
				dns_options.target_domain=input("enter domain: ").replace(" ","")
			elif prompt==2:
				dns_options.wordlist=input("wordlist file: ")
			elif prompt==3:
				try:
					dns_options.threads=int(input("enter threads: "))
				except:
					dns_options.threads=10
			elif prompt==97:
				dns_module.run()
			elif prompt==98:
				dns_module.export()
			elif prompt==99:
				Banner.display_banner=True
				break
			else:
				message.error("invalid input!")

	def run():
		message.info("starting dns enumeration...")
		run=True
		if not dns_options.target_domain:
			message.error("target domain is required!")
			run=False

		if not dns_options.wordlist:
			dns_options.wordlist='wordlist/subdomains.txt'
		
		if not dns_options.threads or dns_options.threads==0:
			dns_options.threads=10

		file = Path(dns_options.wordlist)
		if not file.is_file():
			message.error(f"file {dns_options.wordlist} not found!")
			run=False

		if not validator.domain(dns_options.target_domain):
			message.error("invalid domain!")
			run=False

		if run:
			try:
				dns_options.founds=[]
				with concurrent.futures.ThreadPoolExecutor(max_workers=int(dns_options.threads)) as executor:
					f = open(dns_options.wordlist,'r')
					future_to_url = {executor.submit(dns_module.output,word): word for word in f.read().split("\n")}
		
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
			message.error("dns enumeration not started!")

	def output(line=""):
		try:
			print(f'{line}                             ',end="\r")
			domain_to_request=f'{line}.{dns_options.target_domain}'
			hostname=domain_to_request.replace("..",".")
			res=socket.gethostbyname(hostname)
			print(f'[{c.Green}+{c.Reset}] {c.Green}{hostname}{c.Reset} - [{c.Orange}{res}{c.Reset}]')
			item=[hostname,res]
			dns_options.founds.append(item)
		except:
			pass

	def export():
		message.info(f"exporting results on file {dns_options.target_domain}_dns_enumeration.txt...")
		with open(f'{dns_options.target_domain}_dns_enumeration.txt','w') as f:
			for item in dns_options.founds:
				f.write(f'{item[0]} {item[1]} \n')
		message.success("done!")