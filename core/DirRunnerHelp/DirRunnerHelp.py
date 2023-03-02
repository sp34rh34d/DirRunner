class HELP_MODULE:
	def main():
		print("""

Usage:
    python3 DirRunner.py [module] [args]

Modules
	dir          Uses directory enumeration mode
	dns          Uses DNS subdomain enumeration mode
	file         Uses file enumeration mode
	fingerprint  Uses to detect web technologies
	fuzz         Uses fuzzing mode
	help         Help about any command

Optional args:
  -h, --help              help for DirRunner
  -o, --output            Set filename to save data,
                              txt format :  -o report.txt
                              html format : -o report.html
  -w, --wordlist string   Path to the wordlist

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