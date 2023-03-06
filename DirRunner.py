#!/usr/bin/env python3

import argparse
from core.DirRunnerDNS.DirRunnerDNS import *
from core.DirRunnerFingerprint.DirRunnerFingerprint import *
from core.DirRunnerDir.DirRunnerDirectory import *
from core.DirRunnerFuzz.DirRunnerFuzz import *
from core.DirRunnerFile.DirRunnerFile import *
from core.DirRunnerHelp.DirRunnerHelp import *
from core.DirRunnerVHost.DirRunnerVHost import *

parser = argparse.ArgumentParser(add_help=False)
parser.add_argument('-u','--url',help='Set target URL on Dir enumeration attack')
parser.add_argument('-a','--user-agent',default='DirRunner v1.0',help='Set user-agent')
parser.add_argument('-d','--domain',help='Set target domain on DNS enumeration attack')
parser.add_argument('-w','--wordlist',help='Set wordlist file')
parser.add_argument('-s','--status-code',default='200,301,302,401,403,405',help='Set the status code to print (200,301)')
parser.add_argument('-x','--exts',help='Set target extension files (php,txt,html)')
parser.add_argument('-t','--threads',default='10',help='Set THREADS')
parser.add_argument('-m','--method',default='GET',help='Set method (GET/POST), GET by default.')
parser.add_argument('mode',help='Set attack mode (dir,dns,fingerprint,file,fuzz,help)')
parser.add_argument('-o','--output',help='save data.')
parser.add_argument('-h','--help',action='store_true',help="Skip TLS certificate verification")
parser.add_argument('-k','--no-tls-validation',action='store_false')
parser.add_argument('-U','--username',help="Set username for http auth")
parser.add_argument('-P','--password',help='Set password for http auth')
parser.add_argument('-c','--cookie',help='Set cookies to use for the requests')
parser.add_argument('--timeout',default=15,help='Time each thread waits between requests (15s by default)')
args = parser.parse_args()
ATTACK_MODE=args.mode

if ATTACK_MODE == 'dns':
	DNS_MODULE.main(args)
if ATTACK_MODE == 'fingerprint':
	FINGERPRINT_MODULE.main(args)
if ATTACK_MODE == 'dir':
	DIRECTORY_MODULE.main(args)
if ATTACK_MODE == 'fuzz':
	FUZZ_MODULE.main(args)
if ATTACK_MODE == 'file':
	FILE_MODULE.main(args)
if ATTACK_MODE == 'vhost':
	VHOST_MODULE.main(args)
if ATTACK_MODE == 'help':
	HELP_MODULE.main()