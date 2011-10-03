def debug_mode(request):
	from django.conf import settings
	return {'debug_mode': settings.DEBUG}

def site_version(request):
	from django.conf import settings
	return {'site_version': settings.SITE_VERSION}
