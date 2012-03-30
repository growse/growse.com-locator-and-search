def debug_mode(request):
	from django.conf import settings
	return {'debug_mode': settings.DEBUG}

def site_version(request):
	from django.conf import settings
	return {'site_version': settings.SITE_VERSION}

def date_bools(request):
	import datetime
	return {'april_first': datetime.date.today().month == 4 and datetime.date.today().day == 1}
