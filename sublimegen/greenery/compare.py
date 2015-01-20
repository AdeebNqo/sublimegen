import sys
from greenery.lego import parse
import re

subregex = parse(sys.argv[1])
supregex = parse(sys.argv[2])
s = subregex&(supregex.everythingbut())
if s.empty():
	print("subset")
else:
	print("notsubset")