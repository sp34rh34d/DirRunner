#!/usr/bin/env python3

from core.main import *
import sys,readline
from core.DirRunnerDNS import *
from core.DirRunnerFingerprint import *
from core.DirRunnerDirectory import *
from core.DirRunnerFuzz import *
from core.DirRunnerFile import *
from core.DirRunnerVHost import *


def menu():
	print(f"""[{c.Red}1{c.Reset}] DNS enumeration
[{c.Red}2{c.Reset}] Directories enumeration
[{c.Red}3{c.Reset}] File enumeration
[{c.Red}4{c.Reset}] Virtual host enumeration
[{c.Red}5{c.Reset}] Website fingerprint
[{c.Red}6{c.Reset}] Fuzz
[{c.Red}0{c.Reset}] Exit""")

while True:
	Banner.display()
	menu()
	prompt=''
	try:
		prompt=int(input("👾: "))
	except:
		pass

	if prompt==1:
		dns_module.main()
	elif prompt==2:
		directory_module.main()
	elif prompt==3:
		file_module.main()
	elif prompt==4:
		vhost_module.main()
	elif prompt==5:
		fingerprint_module.main()
	elif prompt==6:
		fuzz_module.main()
	elif prompt==0:
		message.warning("Quitting...")
		sys.exit()
	else:
		message.error("invalid option!")