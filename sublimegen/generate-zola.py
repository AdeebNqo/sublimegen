import rstr
import traceback
import sys
import json
import pprint
filename = sys.argv[1]

with open (filename, "r") as myfile:
    data=myfile.read()
    jsondata = json.loads(data)
    patterns = jsondata["patterns"]
    for pattern in patterns:
    	try:
    		regex = pattern["match"]
    		try:
    			for i in range(5):
    				print(rstr.xeger(regex))
    		except:
    			traceback.print_exc()
    	except KeyError as err:
    		print("")

