
class TerminalColor:
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

class Banner:
	def DirRunnerBanner():
		print(f""" 
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
Coded by:{TerminalColor.Red} Adonis Izaguirre {TerminalColor.Reset} Email:{TerminalColor.Red} adonis.izaguirre@kapa7.com / adons@outlook.com {TerminalColor.Reset}
twitter: {TerminalColor.Red}@AdonsIzaguirre{TerminalColor.Reset}
Welcome to DirRunner v1.0
======================================================================================================""")

	def Help():
		print("""
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
use: python3 DirRunner.py fingerprint -u https://www.domain.com 

""")

	def DirModeBanner(Target,Method,Threads,Codes,UserAgent,Wordlist):
		print(f"""Target: {TerminalColor.Green}{Target}{TerminalColor.Reset}
Method: {TerminalColor.Green}{Method}{TerminalColor.Reset}
Attack mode: {TerminalColor.Green}Directories enumeration{TerminalColor.Reset}
Threads: {TerminalColor.Green}{Threads}{TerminalColor.Reset}
Status code: {TerminalColor.Green}{Codes}{TerminalColor.Reset}
User-agent: {TerminalColor.Green}{UserAgent}{TerminalColor.Reset}
Wordlist file: {TerminalColor.Green}{Wordlist}{TerminalColor.Reset}
======================================================================================================""")

	def DnsModeBanner(Target,Threads,Wordlist):
		print(f"""Target: {TerminalColor.Green}{Target}{TerminalColor.Reset}
Attack mode: {TerminalColor.Green}DNS enumeration{TerminalColor.Reset}
Threads: {TerminalColor.Green}{Threads}{TerminalColor.Reset}
Wordlist file: {TerminalColor.Green}{Wordlist}{TerminalColor.Reset}
======================================================================================================""")

	def FileModeBanner(Target,Threads,Codes,UserAgent,Exts,Wordlist):
		print(f"""Target: {TerminalColor.Green}{Target}{TerminalColor.Reset}
Method: {TerminalColor.Green}GET{TerminalColor.Reset}
Attack mode: {TerminalColor.Green}File enumeration{TerminalColor.Reset}
Threads: {TerminalColor.Green}{Threads}{TerminalColor.Reset}
User-agent: {TerminalColor.Green}{UserAgent}{TerminalColor.Reset}
Extension: {TerminalColor.Green}{Exts}{TerminalColor.Reset}
Wordlist file: {TerminalColor.Green}{Wordlist}{TerminalColor.Reset}
======================================================================================================""")

	def FingerprintModeBanner(Target):
		print(f"""Target: {TerminalColor.Green}{Target}{TerminalColor.Reset}
Attack mode: {TerminalColor.Green}Fingerprint{TerminalColor.Reset}
======================================================================================================""")

class ResponseCode:

	ResponseCode_200=False
	ResponseCode_201=False
	ResponseCode_301=False
	ResponseCode_302=False
	ResponseCode_400=False
	ResponseCode_401=False
	ResponseCode_403=False
	ResponseCode_404=False
	ResponseCode_405=False
	ResponseCode_500=False
	ResponseCode_503=False

	def init(codes):
		for s in codes:
			if s == '200':
				ResponseCode.ResponseCode_200=True
			if s == '201':
				ResponseCode.ResponseCode_201=True
			if s == '301':
				ResponseCode.ResponseCode_301=True
			if s == '302':
				ResponseCode.ResponseCode_302=True
			if s == '400':
				ResponseCode.ResponseCode_400=True
			if s == '401':
				ResponseCode.ResponseCode_401=True
			if s == '403':
				ResponseCode.ResponseCode_403=True
			if s == '404':
				ResponseCode.ResponseCode_404=True
			if s == '405':
				ResponseCode.ResponseCode_405=True
			if s == '500':
				ResponseCode.ResponseCode_500=True
			if s == '503':
				ResponseCode.ResponseCode_503=True

	def ReadResponseCode(ResponseHeaders="",Line="",Method=""):

		if ResponseHeaders.status_code == 200:
			if ResponseCode.ResponseCode_200==True:
				print(f'[{TerminalColor.Green}200{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                        ')
		if ResponseHeaders.status_code == 201:
			if ResponseCode.ResponseCode_201==True:
				print(f'[{TerminalColor.Green}201{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                         ')
		if ResponseHeaders.status_code == 301:
			if ResponseCode.ResponseCode_301==True:
				print(f'[{TerminalColor.Yellow}301{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset} -> [{TerminalColor.Blue}{ResponseHeaders.headers["Location"]}{TerminalColor.Reset}]')
		if ResponseHeaders.status_code == 302:
			if ResponseCode.ResponseCode_302==True:
				print(f'[{TerminalColor.Yellow}302{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset} -> [{TerminalColor.Blue}{ResponseHeaders.headers["Location"]}{TerminalColor.Reset}]')
		if ResponseHeaders.status_code == 400:
			if ResponseCode.ResponseCode_400==True:
				print(f'[{TerminalColor.Blue}400{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
		if ResponseHeaders.status_code == 401:
			if ResponseCode.ResponseCode_401==True:
				print(f'[{TerminalColor.Blue}401{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
		if ResponseHeaders.status_code == 403:
			if ResponseCode.ResponseCode_403==True:
				print(f'[{TerminalColor.Blue}403{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
		if ResponseHeaders.status_code == 404:
			if ResponseCode.ResponseCode_404==True:
				print(f'[{TerminalColor.Blue}404{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
		if ResponseHeaders.status_code == 405:
			if ResponseCode.ResponseCode_405==True:
				print(f'[{TerminalColor.Blue}405{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                          ')
		if ResponseHeaders.status_code == 500:
			if ResponseCode.ResponseCode_500==True:
				print(f'[{TerminalColor.Red}500{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                           ')
		if ResponseHeaders.status_code == 503:
			if ResponseCode.ResponseCode_503==True:
				print(f'[{TerminalColor.Red}503{TerminalColor.Reset}][{TerminalColor.Yellow}{Method}{TerminalColor.Reset}] {TerminalColor.Green}/{Line}{TerminalColor.Reset}                           ')




