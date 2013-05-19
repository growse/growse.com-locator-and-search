# Django settings for growse_com project.
import sys
if not 'runserver' in sys.argv:
    DEBUG = False
else:
    DEBUG = True
TEMPLATE_DEBUG = DEBUG

ADMINS = (
    ('Andrew Rowson', 'andrew@growse.com'),
)
CDN_URL = ('growseres1-growsecom.netdna-ssl.com')

MANAGERS = ADMINS

DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.sqlite3',  # Add 'postgresql_psycopg2', 'postgresql', 'mysql', 'sqlite3' or 'oracle'.
        'NAME': 'growse_com.db',                      # Or path to database file if using sqlite3.
    }
}

# Local time zone for this installation. Choices can be found here:
# http://en.wikipedia.org/wiki/List_of_tz_zones_by_name
# although not all choices may be available on all operating systems.
# On Unix systems, a value of None will cause Django to use the same
# timezone as the operating system.
# If running in a Windows environment this must be set to the same as your
# system time zone.
TIME_ZONE = 'Europe/London'

# Language code for this installation. All choices can be found here:
# http://www.i18nguy.com/unicode/language-identifiers.html
LANGUAGE_CODE = 'en-uk'

SITE_ID = 1

USE_TZ = True

# If you set this to False, Django will make some optimizations so as not
# to load the internationalization machinery.
USE_I18N = False

# If you set this to False, Django will not format dates, numbers and
# calendars according to the current locale
USE_L10N = True

# Absolute path to the directory that holds media.
# Example: "/home/media/media.lawrence.com/"
MEDIA_ROOT = ''

# URL that handles the media served from MEDIA_ROOT. Make sure to use a
# trailing slash if there is a path component (optional in other cases).
# Examples: "http://media.lawrence.com", "http://example.com/media/"
PIPELINE = not DEBUG
if DEBUG:
    STATIC_ROOT = 'static-root'
else:
    STATIC_ROOT = '/var/www/growse.com/res/django-static/www/'
STATICFILES_DIRS = ('static/',)
if DEBUG:
    STATIC_URL = '/static-css/'
else:
    STATIC_URL = '//growseres1-growsecom.netdna-ssl.com/django-static/www/'
if not DEBUG:
    STATICFILES_STORAGE = 'pipeline.storage.PipelineCachedStorage'
PIPELINE_STORAGE = 'pipeline.storage.PipelineFinderStorage'
PIPELINE_CSS = {
    'www': {
        'source_filenames': (
            'css/style.scss',
            'css/wlefi.scss'
        ),
        'output_filename': 'css/www.css',
        'extra_context': {
            'media': 'screen,projection',
        },
    },
}
PIPELINE_JS = {
    'www': {
        'source_filenames': (
            'js/jquery-*.min.js',
            'js/jquery.*.js',
            'js/scripts.js',
        ),
        'output_filename': 'js/www.js',
    }
}
PIPELINE_COMPILERS = (
    'pipeline_compass.compiler.CompassCompiler',
)
PIPELINE_CSS_COMPRESSOR = None
PIPELINE_JS_COMPRESSOR = None
PIPELINE_DISABLE_WRAPPER = True


# URL prefix for admin media -- CSS, JavaScript and images. Make sure to use a
# trailing slash.
# Examples: "http://foo.com/media/", "/media/".
#ADMIN_MEDIA_PREFIX = '/media/'

# Make this unique, and don't share it with anybody.
SECRET_KEY = ''

# List of callables that know how to import templates from various sources.
TEMPLATE_LOADERS = (
    'django.template.loaders.filesystem.Loader',
    'django.template.loaders.app_directories.Loader',
)


TEMPLATE_CONTEXT_PROCESSORS = (
    'growse_com.blog.context_processors.debug_mode',
    'growse_com.blog.context_processors.site_version',
    'growse_com.blog.context_processors.date_bools',
    'django.contrib.auth.context_processors.auth',
    'django.core.context_processors.request',
)

MIDDLEWARE_CLASSES = (
    'django.middleware.common.CommonMiddleware',
    'djangosecure.middleware.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'django.contrib.messages.middleware.MessageMiddleware',
    'django.middleware.gzip.GZipMiddleware',
)


if DEBUG:
    SECURE_SSL_REDIRECT = False
    SECURE_FRAME_DENY = True
    SECURE_HSTS_SECONDS = 300
    SECURE_HSTS_INCLUDE_SUBDOMAINS = False
    SECURE_CONTENT_TYPE_NOSNIFF = True
    SECURE_BROWSER_XSS_FILTER = True
    SESSION_COOKIE_SECURE = True
else:
    SECURE_SSL_REDIRECT = False

ROOT_URLCONF = 'growse_com.urls'

TEMPLATE_DIRS = (
    # Put strings here, like "/home/html/django_templates" or "C:/www/django/templates".
    # Always use forward slashes, even on Windows.
    #! Don't forget to use absolute paths, not relative paths.
    #"/home/growse/django/templates",
)

INSTALLED_APPS = (
    'django.contrib.staticfiles',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.sites',
    'django.contrib.messages',
    'django.contrib.sitemaps',
    'growse_com.blog',
    # Uncomment the next line to enable the admin:
    'django.contrib.admin',
    # Uncomment the next line to enable admin documentation:
    # 'django.contrib.admindocs',
    'djangosecure',
    'pipeline',
)

#CACHES = {
#		'default': {
#			'BACKEND': 'django.core.cache.backends.filebased.FileBasedCache',
#			'LOCATION': '/var/tmp/django_cache',
#		}
#	}
#
#CACHE_MIDDLEWARE_KEY_PREFIX='growse_com'
#CACHE_MIDDLEWARE_SECONDS=300

FORCE_SCRIPT_NAME = ''

try:
    from local_settings import *
except ImportError:
    pass
