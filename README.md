# DirRunner
DNS, Directories and file enumeration and fingerprint tool

install requeriments
```
pip3 install -r requeriments.txt
```

optional arguments:
```
  -u: set target url
  -d: set target domain
  -a: set user-agent
  -x: set target extensions files (php,txt,html)
  -s: set the status code to print (200,301)
  -w: set wordlist
  -t: set threads
  -m: set method (GET/POST/DELETE/PUT/HEAD/OPTION), GET by default.
  -h: show this message
```
DNS Enumeration mode:
```
  python3 DirRunner.py dns -d domain.com -w wordlist.txt
```
Dir enumeration mode:
```
  python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt
```
Print only codes 200 and 301
```
  python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -s 200,301
```
file discovery
```
  python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x php,txt
```
fingerprint for get web technology
```
  python3 DirRunner.py fingerprint -u https://www.domain.com
```
