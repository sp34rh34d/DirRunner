from core.main import *
import requests, string, random, urllib3, json, concurrent.futures,json
from pathlib import Path
from datetime import datetime
urllib3.disable_warnings()
from tabulate import tabulate

class fuzz_option:
	module_name="Fuzz enumeration"
	target_url=""
	method='get'
	wordlist="wordlist/directory.txt"
	user_agent="DirRunner v1.2"
	threads=10
	status_code=[200,301,302]
	tls_validation=True
	follow_redirects=False
	custom_headers={}
#	custom_header_value=""
	timeout=10
	cookie=""
	username=""
	password=""
	custom_post_parameter=""#{}
	valid_http_methods=['get','post','head','delete','put']
	fuzz=False
	founds=[]

class fuzz_module:
	def menu():
		menu=[]
		module=[f'{c.Blue}#{c.Reset}','Module',f'{c.Green}{fuzz_option.module_name}{c.Reset}']
		menu.append([f'{c.Red}1{c.Reset}','Target',f'{c.Green}{fuzz_option.target_url}{c.Reset} ({c.Orange}required{c.Reset})'])
		menu.append([f'{c.Red}2{c.Reset}','Method',f'{c.Green}{fuzz_option.method}{c.Reset}'])
		menu.append([f"{c.Red}3{c.Reset}", "Wordlist" ,f"{c.Green}{fuzz_option.wordlist}{c.Reset}"])
		menu.append([f"{c.Red}4{c.Reset}", "User agent" ,f"{c.Green}{fuzz_option.user_agent}{c.Reset}"])
		menu.append([f"{c.Red}5{c.Reset}", "Threads" ,f"{c.Green}{fuzz_option.threads}{c.Reset}"])
		menu.append([f"{c.Red}6{c.Reset}", "Show status code" ,f"{c.Green}{fuzz_option.status_code}{c.Reset}"])
		menu.append([f"{c.Red}7{c.Reset}", "TLS validation" ,f"{c.Green}{fuzz_option.tls_validation}{c.Reset}"])
		menu.append([f"{c.Red}8{c.Reset}", "Follow redirect" ,f"{c.Green}{fuzz_option.follow_redirects}{c.Reset}"])
		menu.append([f"{c.Red}9{c.Reset}", "TimeOut" ,f"{c.Green}{fuzz_option.timeout}{c.Reset}"])
		menu.append([f"{c.Red}10{c.Reset}", "Cookies" ,f"{c.Green}{fuzz_option.cookie}{c.Reset}"])
		menu.append([f"{c.Red}11{c.Reset}", "Username" ,f"{c.Green}{fuzz_option.username}{c.Reset}"])
		menu.append([f"{c.Red}12{c.Reset}", "Password" ,f"{c.Green}{fuzz_option.password}{c.Reset}"])
		menu.append([f"{c.Red}13{c.Reset}", "Custom header" ,f"{c.Green}{fuzz_option.custom_headers}{c.Reset}"])
		menu.append([f"{c.Red}14{c.Reset}", "Custom POST parameters" ,f"{c.Green}{fuzz_option.custom_post_parameter}{c.Reset}"])
		menu.append([f'{c.Red}97{c.Reset}','Run',''])
		menu.append([f'{c.Red}98{c.Reset}','Export results',''])
		menu.append([f'{c.Red}99{c.Reset}','Back',''])
		print(tabulate(menu, module,  tablefmt="grid",numalign="left"))

	def main():
		while True:
			Banner.display_banner=True
			Banner.display()
			fuzz_module.menu()
			prompt=''
			try:
				prompt=int(input("🫨 : "))
			except:
				pass

			if prompt==99:
				Banner.display_banner=True
				break
			elif prompt==1:
				fuzz_option.target_url=input("enter target url: ").replace(" ","")
			elif prompt==2:
				try:
					method=input("enter http method: ")
					fuzz_option.method=method
					if not method:
						fuzz_option.method='get'
				except:
					fuzz_option.method='get'
			elif prompt==3:
				fuzz_option.wordlist=input("enter wordlist path: ")
			elif prompt==4:
				fuzz_option.user_agent=input("enter user agent: ")
			elif prompt==5:
				try:
					fuzz_option.threads=int(input("enter threads: "))
				except:
					fuzz_option.threads=10
			elif prompt==6:
				try:
					fuzz_option.status_code=input("enter status code to display (200,301): ").split(',')
				except:
					fuzz_option.status_code=[200,301,302]
			elif prompt==7:
				validate=input("do you want to validate tls (y/n): ")
				fuzz_option.tls_validation=True
				if validate=='n':
					fuzz_option.tls_validation=False
			elif prompt==8:
				validate=input("follow redirects (y/n): ")
				fuzz_option.follow_redirects=True
				if validate=='n':
					fuzz_option.follow_redirects=False

			elif prompt==9:
				try:
					fuzz_option.timeout=int(input("enter timeout (seconds): "))
				except:
					fuzz_option.timeout=10
			elif prompt==10:
				fuzz_option.cookie=input("enter cookies: ")
			elif prompt==11:
				fuzz_option.username=input("enter username: ")
			elif prompt==12:
				fuzz_option.password=input("enter password: ")
			elif prompt==13:
				# fuzz_option.custom_header_name=input("enter custom header name: ")
				# fuzz_option.custom_header_value=input("enter custom header value: ")
				while True:
					header_name=input("enter header name: ")
					header_value=input("enter header value: ")
					fuzz_option.custom_headers[header_name]=header_value
					validate=input("do you want to add another header (y/n): ")
					if validate=='n':
						break

			elif prompt==14:
				fuzz_option.custom_post_parameter=input('enter POST parameters (json format {"key":"FUZZ"} ): ')
				if fuzz_option.custom_post_parameter:
					fuzz_option.method='post'
				# while True:
				# 	parameter_name=input("enter parameter name: ")
				# 	parameter_value=input("enter parameter value: ")
				# 	fuzz_option.custom_post_parameter[parameter_name]=parameter_value
				# 	validate=input("do you want to add another parameter (y/n): ")
				# 	if validate=='n':
				# 		break

			elif prompt==97:
				fuzz_module.run()
			elif prompt==98:
				fuzz_module.export()
			else:
				message.error("invalid input!")


	def run():
		message.info("starting fuzz enumeration...")
		run=True
		if not fuzz_option.target_url or not validator.url(fuzz_option.target_url):
			message.error("invalid target url!")
			run=False
			return
		
		if not fuzz_option.method.lower() in fuzz_option.valid_http_methods:
			message.error("invalid http method detected!")
			run=False
			return
		
		for code in fuzz_option.status_code:
			if not int(code) >199 or not int(code) <600: 
				message.error("invalid status code detected!")
				run=False
				return
		if not fuzz_option.wordlist:
			fuzz_option.wordlist='wordlist/directory.txt'

		file = Path(fuzz_option.wordlist)
		if not file.is_file():
			message.error(f"file {fuzz_option.wordlist} not found!")
			run=False
			return

		if not fuzz_module.validate_connection():
			message.error("connection error detected!")
			return

		for code in fuzz_option.status_code:
			if fuzz_module.check_non_existing_url(int(code)):
				run=False
				return
			message.info("done!")
		
		if fuzz_option.custom_post_parameter:
			if not fuzz_module.is_valid_json():
				message.error("invalid json format")
				run=False
				return

		message.info("looking for 'FUZZ' parameter...")
		fuzz_module.validate_fuzz_parameter()
		if not fuzz_option.fuzz:
			message.error("'FUZZ' word no detected, please put 'FUZZ' word on target url, user-agent, username, password, cookie, post parameters or custom header name/value")
			run=False
			return
		
		if run:
			try:
				fuzz_option.founds=[]
				message.info("running fuzz enumeration...")
				with concurrent.futures.ThreadPoolExecutor(max_workers=int(fuzz_option.threads)) as executor:
					f = open(fuzz_option.wordlist,'r')
					future_to_url = {executor.submit(fuzz_module.request,word): word for word in f.read().split("\n")}
		
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
			message.error("fuzz enumeration not started!")

	def validate_fuzz_parameter():
		if "FUZZ" in fuzz_option.target_url or "FUZZ" in fuzz_option.custom_post_parameter or "FUZZ" in fuzz_option.user_agent or "FUZZ" in fuzz_option.cookie or "FUZZ" in fuzz_option.username or "FUZZ" in fuzz_option.password:
			fuzz_option.fuzz=True

		for key in fuzz_option.custom_headers:
			if "FUZZ" in fuzz_option.custom_headers[key]:
				fuzz_option.fuzz=True
		
	def is_valid_json():
		try:
			json.loads(fuzz_option.custom_post_parameter)
			return True
		except json.decoder.JSONDecodeError:
			return False

	def validate_connection():
		try:
			message.info(f"checking connection for target {fuzz_option.target_url}")
			custom_headers={
				"User-Agent":fuzz_option.user_agent,
				"cookie":fuzz_option.cookie,
			}

			auth = None
			if fuzz_option.username and fuzz_option.password:
				auth = (fuzz_option.username, fuzz_option.password)

			custom_request = requests.get(
				fuzz_option.target_url,
				auth=auth,
				headers=custom_headers,
				allow_redirects=fuzz_option.follow_redirects,
				timeout=fuzz_option.timeout,
				verify=fuzz_option.tls_validation
				)
			message.info("done!")
			return True
		except requests.exceptions.Timeout:
			message.error(f"timeout for {fuzz_option.target_url}")
			return False
		except requests.exceptions.SSLError:
			message.error("ssl verification error!")
			return False
		except requests.exceptions.TooManyRedirects:
			message.error(f"too may redirect for {fuzz_option.target_url}")
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
		random_directory=fuzz_module.RandomStrings()
		url=f"{fuzz_option.target_url}/{random_directory}"
		message.info(f"checking for non existing directory {url} for status code {code}")

		status_code,headers=fuzz_module.send_http_request(url,random_directory)
		if status_code is not None:
			if status_code==code:
				message.warning(f"we got a valid http {fuzz_option.method} response from non existing directory, please exclude {code}")
				return True
			else:
				return False

	def send_http_request(url,line):
		custom_headers={
			"User-Agent":fuzz_option.user_agent.replace("FUZZ",line),
			"cookie":fuzz_option.cookie.replace("FUZZ",line)
		}

		for header in fuzz_option.custom_headers:
			if "FUZZ" in fuzz_option.custom_headers[header]:
				custom_headers[header]=line

		auth = None
		if fuzz_option.username and fuzz_option.password:
			auth = (fuzz_option.username.replace("FUZZ",line), fuzz_option.password.replace("FUZZ",line))

		try:
				if fuzz_option.method.lower()=="get":
					custom_request=requests.get(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=fuzz_option.follow_redirects,
						timeout=fuzz_option.timeout,
						verify=fuzz_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers
				
				if fuzz_option.method.lower()=="post":

					post_values=json.loads(fuzz_option.custom_post_parameter.replace("FUZZ",line))
					custom_request=requests.post(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=fuzz_option.follow_redirects,
						timeout=fuzz_option.timeout,
						verify=fuzz_option.tls_validation,
						json=post_values
					)
					return custom_request.status_code,custom_request.headers
				
				if fuzz_option.method.lower()=="head":
					custom_request=requests.head(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=fuzz_option.follow_redirects,
						timeout=fuzz_option.timeout,
						verify=fuzz_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers
				
				if fuzz_option.method.lower()=="delete":
					custom_request=requests.put(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=fuzz_option.follow_redirects,
						timeout=fuzz_option.timeout,
						verify=fuzz_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers
				
				if fuzz_option.method.lower()=="put":
					custom_request=requests.put(
						url,
						auth=auth,
						headers=custom_headers,
						allow_redirects=fuzz_option.follow_redirects,
						timeout=fuzz_option.timeout,
						verify=fuzz_option.tls_validation
					)
					return custom_request.status_code,custom_request.headers

		except requests.exceptions.Timeout:
			message.error(f"connection timeout for {url}")
			return None, None
		# except requests.exceptions.RequestException as e:
		# 	message.error(e)
		# 	return None,None
		except:
			return None, None

	def request(line):
		print(f'{line}                             ',end="\r")
		url=f"{fuzz_option.target_url}".replace("FUZZ",line)
		status_code,headers=fuzz_module.send_http_request(url,line)
		if status_code is not None and status_code in fuzz_option.status_code:
			fuzz_module.output(status_code,headers,line)

	def output(code,header,line):
		if code >199 and code <300:
			print(f"[{c.Green}{code}{c.Reset}] {line} [size: {header['Content-Length']}]")
		elif code >299 and code <400:
			print(f"[{c.Orange}{code}{c.Reset}] {line} [size: {header['Content-Length']}] -> [{header['Location']}]")
		elif code >399 and code <500:
			print(f"[{c.Blue}{code}{c.Reset}] {line} [size: {header['Content-Length']}]")
		elif code >499 and code <600:
			print(f"[{c.Red}{code}{c.Reset}] {line} [size: {header['Content-Length']}]")

		try:
			item=[code,fuzz_option.method,f"{fuzz_option.target_url} {line}",header["Content-Length"],header["Location"]]
		except:
			item=[code,fuzz_option.method,f"{fuzz_option.target_url} {line}",header["Content-Length"],'']
		fuzz_option.founds.append(item)

	def export():
		random_name=fuzz_module.RandomStrings()
		message.info(f"exporting results on file {random_name}_fuzz.txt...")
		with open(f'{random_name}_fuzz.txt','w') as f:
			for item in fuzz_option.founds:
				f.write(f'{item[0]} {item[1]} {item[2]} {item[3]} {item[4]}\n')
		message.success("done!")






