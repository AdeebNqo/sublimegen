import rstr
import traceback

regex = "{(a|b)}"
try:
	for i in range(5):
		print(rstr.xeger(regex))
except:
	traceback.print_exc()

