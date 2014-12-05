import json
import plistlib
import sys

inputf = sys.argv[1]
outputf = sys.argv[2]

jsondata = open(inputf)
data = json.load(jsondata)
plistlib.writePlist(data, outputf)
