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
def pfn_abs(*args):
	if len(args) < 1:
		raise ArgcountError('too few arguments for function abs')
	if 0 != args[0]:
		raise UnmatchedError('unmatched')
	return "zero"

def pfn_abs_1(*args):
	if len(args) < 1:
		raise ArgcountError('too few arguments for function abs_1')
	x = args[0]
	return x

print(__pfn_call([pfn_abs,pfn_abs_1], [0]))
print(__pfn_call([pfn_abs,pfn_abs_1], [9]))
