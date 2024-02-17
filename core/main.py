import re
import ipaddress

class c:
	Black = '\033[30m'
	Red = '\033[31m'
	Green = '\033[32m'
	Orange = '\033[33m'
	Blue = '\033[34m'
	Purple = '\033[35m'
	Reset = '\033[0m'
	Cyan = '\033[36m'
	LightGrey = '\033[37m'
	DarkGrey = '\033[90m'
	LightRed = '\033[91m'
	LightGreen = '\033[92m'
	Yellow = '\033[93m'
	LightBlue = '\033[94m'
	Pink = '\033[95m'
	LightCyan = '\033[96m'

class message():
	def error(msg):
		print(f'{c.Red}{msg}{c.Reset}')

	def success(msg):
		print(f'{c.Green}{msg}{c.Reset}')
	
	def warning(msg):
		print(f'{c.Orange}{msg}{c.Reset}')

	def info(msg):
		print(f'{c.Blue}{msg}{c.Reset}')


class Banner:
	display_banner=True

	def display():
		if Banner.display_banner:
			Banner.DirRunnerBanner()
			Banner.display_banner=False

	def DirRunnerBanner():
		print(f""" 
   ___  _     ___                        
  / _ \(_)___/ _ \__ _____  ___  ___ ____
 / // / / __/ , _/ // / _ \/ _ \/ -_) __/
/____/_/_/ /_/|_|\_,_/_//_/_//_/\__/_/ 
	                                
Coded by:{c.Red} sp34rh34d {c.Reset}
twitter: {c.Red}@AdonsIzaguirre{c.Reset}
Welcome to DirRunner v1.2 [{c.Green}https://github.com/sp34rh34d/DirRunner{c.Reset}]
======================================================================================================""")
		

class validator:
    
    def domain(value):
        try:
            ipaddress.ip_address(value)
            return False
        except ValueError:
            domain_pattern = re.compile(r"^(?!-)[A-Za-z0-9-]{1,63}(?<!-)(\.[A-Za-z0-9-]{1,63})*$")
            return bool(domain_pattern.match(value))

    def ip_address(value):
        try:
            ipaddress.ip_address(value)
            return True
        except ValueError:
            return False

    def url(value):
        try:
            url_pattern = re.compile(
                r'^(https?|ftp):\/\/'  # Scheme (http, https, or ftp)
                r'(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+(?:[A-Z]{2,6}\.?|[A-Z0-9-]{2,}\.?)|'  # Domain
                r'localhost|'  # localhost
                r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|'  # IP address
                r'\[?[A-F0-9]*:[A-F0-9:]+\]?)'  # IPv6 address
                r'(?::\d+)?'  # Port
                r'(?:/?|[/?]\S+)$', re.IGNORECASE
            )

            return bool(url_pattern.match(value))
        except:
            return False

    def hostname(value):
        try:
            ipaddress.ip_address(value)
            return False
        except:
            hostname_pattern = re.compile(r'^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$')
            return bool(hostname_pattern.match(value))

