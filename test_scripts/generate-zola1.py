#
# Copyright 2015 Zola Mahlaza <adeebnqo@gmail.com>
# 19 January 2015
#
# This is a simple script for generating strings
# which the regex provided as a string. It's manily
# used for testing.

import rstr
import traceback

regex = "{(a|b)}"
try:
	for i in range(5):
		print(rstr.xeger(regex))
except:
	traceback.print_exc()

