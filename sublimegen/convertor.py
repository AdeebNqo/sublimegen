import json
import plistlib
import sys
import re



#The above methods and variables were written by Martijn Pieters(http://stackoverflow.com/a/15198886/1984350)

inputf = sys.argv[1]
outputf = sys.argv[2]

with open (inputf, "r") as myfile:
    data=myfile.read()
    jsondata = json.loads(data)
    plistlib.writePlist(jsondata, outputf)
