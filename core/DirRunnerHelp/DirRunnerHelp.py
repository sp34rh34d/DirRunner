class HELP_MODULE:
	def main():
		print("""
Modules
	dir         Uses directory enumeration mode
	dns         Uses DNS subdomain enumeration mode
	file        Uses file enumeration mode
	fingerprint Uses to detect web technologies
	fuzz        Uses fuzzing mode
	help        Help about any command

Optional args
	-u,--url: set target url
	-d,--domain: set target domain
	-a,--user-agent: set user-agent 'DirRunner v1.0' by default
	-x,--exts: set target extensions files (php,txt,html)
	-s,--status-code: set the status code to print (200,301)
	-w,--wordlist: set wordlist file
	-t,--threads: set threads
	-m,--method: set method (GET/POST/DELETE/OPTION/PUT/HEAD) for requests, GET by default.
	-h,--help: show this message

Generate outputs files
-o,--output: set filename to save data,
	txt format :  -o report.txt
	html format : -o report.html

Examples:
	dns enumeration
	use: python3 DirRunner.py dns -d domain.com -w wordlist.txt

	dns enumeration with TXT output
	use: python3 DirRunner.py dns -d domain.com -w wordlist.txt -o dns-output.txt

	dir enumeration
	use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt

	print only status code 200 and 301
	use: python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -s 200,301

	file detection by extensions
	use: python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x php,txt

	fingerprint
	use: python3 DirRunner.py fingerprint -u https://www.domain.com 

	fuzz 
	use: python3 DirRunner.py fuzz -u https://www.domain.com/FUZZ -w wordlist.txt

""")