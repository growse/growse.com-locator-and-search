def debug_mode(request):
	from django.conf import settings
	from django.core.cache import cache 
	return {'debug_mode': settings.DEBUG,'cache':cache}

def site_version(request):
	from django.conf import settings
	return {
			'cdn_url': settings.CDN_URL
			}

def date_bools(request):
	import datetime
	return {'april_first': datetime.date.today().month == 4 and datetime.date.today().day == 1}
