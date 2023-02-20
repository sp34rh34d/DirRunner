# DirRunner
DNS, Directories and file enumeration and fingerprint tool

<h3>Modules</h3>
<li>dir: directories enumeration mode</li>
<li>dns: dns enumeration mode</li>
<li>file: file enumeration mode by extension</li>
<li>fingerprint: web technology detection mode</li>
<br>

install requirements
```
pip3 install -r requirements.txt
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

https://user-images.githubusercontent.com/94752464/219988038-8a548369-cef0-4303-a0b5-4fdc4ce23657.mov

<br>

<img width="866" alt="Captura de Pantalla 2023-02-16 a la(s) 17 50 01" src="https://user-images.githubusercontent.com/94752464/219514018-c0d0ae15-f366-42e3-8895-1ace942e5b3f.png">


Dir enumeration mode:
```
  python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt
```

https://user-images.githubusercontent.com/94752464/219988070-b232f24f-fd96-4f4e-bd4b-b7cd4c29255d.mov


Print only codes 200 and 301
```
  python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -s 200,301
```
file discovery
```
  python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x php,txt
```

https://user-images.githubusercontent.com/94752464/219988112-28eaa466-d7ac-4fb7-9055-b47a71f6dad4.mov

<br>

fingerprint for get web technology
```
  python3 DirRunner.py fingerprint -u https://www.domain.com
```
https://user-images.githubusercontent.com/94752464/219988133-8c197c84-5a41-40fa-971a-b4b1887fbbd6.mov

<br>

<img width="828" alt="Captura de Pantalla 2023-02-16 a la(s) 17 54 23" src="https://user-images.githubusercontent.com/94752464/219514378-771db89b-01e7-4cde-abab-04ef6cf379f9.png">

