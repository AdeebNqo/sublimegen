import json
import plistlib
import sys
import re


invalid_escape = re.compile(r'\\[0-7]{1,3}')  # up to 3 digits for byte values up to FF

def replace_with_byte(match):
    return chr(int(match.group(0)[1:], 8))

def repair(brokenjson):
    return invalid_escape.sub(replace_with_byte, brokenjson)

#The above methods and variables were written by Martijn Pieters(http://stackoverflow.com/a/15198886/1984350)

inputf = sys.argv[1]
outputf = sys.argv[2]

with open (inputf, "r") as myfile:
    data=myfile.read()
    jsondata = json.loads(repair(data))
    plistlib.writePlist(jsondata, outputf)
