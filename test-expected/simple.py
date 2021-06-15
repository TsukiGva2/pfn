# this code was auto generated by pfn

class UnmatchedError(Exception):
	pass

class ArgcountError(Exception):
	pass

def __pfn_call(p, args):
	result=None
	broke=False
	for f in p:
		try:
			result=f(*args)
		except (UnmatchedError, ArgcountError):
			continue
		broke=True

		break
	if not broke:
		raise Exception('no matching function')

	return result
def pfn_f(*args):
	if len(args) < 2:
		raise ArgcountError('too few arguments for function f')
	x = args[0]
	y = args[1]
	return (x+y)

a=2
print(__pfn_call([pfn_f], [a,5]))