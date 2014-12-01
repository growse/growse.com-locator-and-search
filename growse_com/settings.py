# Django settings for growse_com project.
import sys

if 'runserver' not in sys.argv:
    DEBUG = False
else:
    DEBUG = True

TEMPLATE_DEBUG = DEBUG

ADMINS = (
    ('Andrew Rowson', 'andrew@growse.com'),
)
CDN_URL = 'growseres1-growsecom.netdna-ssl.com'

MANAGERS = ADMINS

DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql_psycopg2',
        'NAME': 'www_growse_com',
        'USER': 'www_growse_com',
        'PASSWORD': 'password',
        'HOST': '',
        'CONN_MAX_AGE': None
    }
}

ALLOWED_HOSTS = ['www.growse.com']

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
if DEBUG:
    STATIC_ROOT = 'static-root'
else:
    STATIC_ROOT = '/var/www/res.growse.com/django-static/www/'

STATICFILES_DIRS = ('static/',
                    'lib/python2.7/site-packages/suit/static/')

STATICFILES_FINDERS = (
        'django.contrib.staticfiles.finders.FileSystemFinder',
        'django.contrib.staticfiles.finders.AppDirectoriesFinder',
        'pipeline.finders.PipelineFinder',
)

if DEBUG:
    STATIC_URL = '/static/'
else:
    STATIC_URL = '//growseres1-growsecom.netdna-ssl.com/django-static/www/'

STATICFILES_STORAGE = 'pipeline.storage.PipelineCachedStorage'
PIPELINE_ENABLED = not DEBUG
PIPELINE_STORAGE = 'pipeline.storage.PipelineFinderStorage'
PIPELINE_CSS = {
    'www': {
        'source_filenames': (
            'css/nanoscroller.css',
            'css/style.scss',
            'css/solarizeddark.scss'
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
    },
    'd3': {
        'source_filenames': (
            'js/d3.js',
            'js/topojson.v1.min.js',
        ),
        'output_filename': 'js/d3-min.js',
    }

}
PIPELINE_COMPILERS = (
    'pipeline_compass.compiler.CompassCompiler',
)
PIPELINE_CSS_COMPRESSOR = 'pipeline.compressors.yui.YUICompressor'
PIPELINE_JS_COMPRESSOR = 'pipeline.compressors.yui.YUICompressor'
PIPELINE_DISABLE_WRAPPER = True

CACHES = {
    'default': {
        'BACKEND': 'django.core.cache.backends.dummy.DummyCache',
    }
}

# URL prefix for admin media -- CSS, JavaScript and images. Make sure to use a
# trailing slash.
# Examples: "http://foo.com/media/", "/media/".
# ADMIN_MEDIA_PREFIX = '/media/'

# Make this unique, and don't share it with anybody.
SECRET_KEY = 'INSECURE_DEFAULT'

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
    'django.middleware.gzip.GZipMiddleware',
    'debug_toolbar.middleware.DebugToolbarMiddleware',
    'growse_com.blog.middleware.SmartUpdateCacheMiddleware',
    'django.middleware.common.CommonMiddleware',
    'djangosecure.middleware.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'django.contrib.messages.middleware.MessageMiddleware',
    'django.middleware.cache.FetchFromCacheMiddleware',
)

if not DEBUG:
    SECURE_SSL_REDIRECT = True
    SECURE_FRAME_DENY = True
    SECURE_HSTS_SECONDS = 31536000
    SECURE_HSTS_INCLUDE_SUBDOMAINS = True
    SECURE_CONTENT_TYPE_NOSNIFF = True
    SECURE_BROWSER_XSS_FILTER = True
    SESSION_COOKIE_SECURE = True
else:
    SECURE_SSL_REDIRECT = False

ROOT_URLCONF = 'growse_com.urls'

TEMPLATE_DIRS = ()

INSTALLED_APPS = (
    'django.contrib.staticfiles',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.sites',
    'django.contrib.messages',
    'django.contrib.sitemaps',
    'django.contrib.humanize',
    'debug_toolbar',
    'growse_com.blog',
    'suit',
    'django.contrib.admin',
    'django_extensions',
    'djangosecure',
    'pipeline',
)

INTERNAL_IPS = ('127.0.0.1',)

EMAIL_BACKEND = 'django.core.mail.backends.filebased.EmailBackend'
EMAIL_FILE_PATH = '/tmp/growse.com-mail'

DEFAULT_FROM_EMAIL = 'blog@growse.com'

FORCE_SCRIPT_NAME = ''

GEOCODE_API_URL = 'http://api.geonames.org/findNearbyPlaceNameJSON?formatted=false&lat={}&lng={}&cities=cities15000&username=growse'

TEST_RUNNER = 'django.test.runner.DiscoverRunner'
WSGI_APPLICATION = "growse_com.wsgi.application"
try:
    from local_settings import *
except ImportError:
    pass
