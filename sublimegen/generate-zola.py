#
# Copyright 2015 Zola Mahlaza <adeebnqo@gmail.com>
# 19 January 2015
#
# This is a script used for testing. It generates
# strings which match the regexes in a sublime text
# syntax highlighting file which is in json format.
# It will also produce the first regex in the list
# which the string matches.
#
import rstr
import traceback
import sys
import json
import pprint
filename = sys.argv[1]

with open (filename, "r") as myfile:
	ws = re.compile(r'\s+')
	data=myfile.read()
	jsondata = json.loads(data)
	patterns = jsondata["patterns"]
	for pattern in patterns:
		if not pattern:
			pass
		else:
			regex = pattern["match"]
			for i in range(1):
				val = rstr.xeger(regex)
				print(val)
				#print(re.sub(ws, '', val))
				for ipattern in patterns:
					if not ipattern:
						pass
					else:
						compiledregex = re.compile(ipattern["match"])
						if compiledregex.match(val):
							print("first match: "+ipattern["match"])
							break
		print('\n')

