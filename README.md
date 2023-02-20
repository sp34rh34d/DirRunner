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
![dns](https://user-images.githubusercontent.com/94752464/219993102-e13e4604-d90f-40eb-88aa-9f6f20e94b9c.gif)

<br>

<img width="866" alt="Captura de Pantalla 2023-02-16 a la(s) 17 50 01" src="https://user-images.githubusercontent.com/94752464/219514018-c0d0ae15-f366-42e3-8895-1ace942e5b3f.png">


Dir enumeration mode:
```
  python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt
```

![dir](https://user-images.githubusercontent.com/94752464/219993126-0918fb64-fd25-4179-9264-aa1a0e281bcd.gif)


Print only codes 200 and 301
```
  python3 DirRunner.py dir -u https://www.domain.com/ -w wordlist.txt -s 200,301
```
file discovery
```
  python3 DirRunner.py file -u https://www.domain.com/ -w wordlist.txt -x php,txt
```

![file](https://user-images.githubusercontent.com/94752464/219993153-55382e05-51d3-453c-b8c2-7d5f6db68df3.gif)

<br>

fingerprint for get web technology
```
  python3 DirRunner.py fingerprint -u https://www.domain.com
```

![fingerprint](https://user-images.githubusercontent.com/94752464/219993178-23034155-7ebf-4167-a20a-0752b51394a8.gif)

<br>

<img width="828" alt="Captura de Pantalla 2023-02-16 a la(s) 17 54 23" src="https://user-images.githubusercontent.com/94752464/219514378-771db89b-01e7-4cde-abab-04ef6cf379f9.png">

