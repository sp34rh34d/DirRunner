# DirRunner
DNS, Directories,virtual host, file enumeration and fingerprint tool

<h3>Modules</h3>
<li>dir         - Uses directory enumeration mode</li>
<li>dns         - Uses DNS subdomain enumeration mode</li>
<li>vhost       - Uses Virtual Host enumeration mode</li>
<li>file        - Uses file enumeration mode</li>
<li>fingerprint - Uses to detect web technologies</li>
<li>fuzz        - Uses fuzzing mode</li>
<li>help        - Help about any command</li>
<br>

install requirements
```
pip3 install -r requirements.txt
```

show module help menu
```
 python3 DirRunner.py [module] -h 
 example: python3 DirRunner.py dir -h
```
optional arguments:
```
  -u, --url                 set target url
  -d, --domain              set target domain
  -a, --user-agent          set user-agent 'DirRunner v1.0' by default
  -x, --exts                set target extensions files (php,txt,html)
  -s, --status-code         set the status code to print (200,301)
  -w, --wordlist            set wordlist file
  -t, --threads             set threads
  -m, --method              set method (GET/POST/DELETE/OPTION/PUT/HEAD) for requests, GET by default.
  -h, --help                show this message
  -c, --cookie              set cookies to use for the requests
  -k, --no-tls-validation   skip TLS certificate verification
  -P, --password            Password for Basic Auth
  -U, --username            Username for Basic Auth
      --timeout             HTTP Timeout (default 15s)
  -o, --output              set filename to save data
                                 txt format  -o report.txt
                                 html format -o report.html
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

Virtual host enumeration mode:
```
  python3 DirRunner.py vhost -d domain.thm -u http://10.10.10.10 -t 30
```
![ezgif com-video-to-gif](https://user-images.githubusercontent.com/94752464/223283537-7f30d56b-27ca-4862-8278-4a79fa750758.gif)


Fuzz enumeration mode:
```
  python3 DirRunner.py fuzz -u https://www.domain.com/FUZZ -w wordlist.txt
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

