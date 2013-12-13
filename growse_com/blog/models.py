# This is an auto-generated Django model module.
# You'll have to do the following manually to clean this up:
#     * Rearrange models' order
#     * Make sure each model has one field with primary_key=True
# Feel free to rename the models, but don't rename db_table values or field names.
#
# Also note: You'll have to insert the output of 'django-admin.py sqlcustom [appname]'
# into your database.
from django.core.mail import send_mail

from django.db import models
from django.utils.html import strip_tags
import datetime
from growse_com import settings
import re
import markdown
from django.core.cache import cache
import requests


class Article(models.Model):
    id = models.AutoField(primary_key=True)
    datestamp = models.DateTimeField(null=True, auto_now_add=True)
    title = models.CharField(max_length=255)
    shorttitle = models.CharField(unique=True, max_length=255)
    description = models.TextField(null=True)
    markdown = models.TextField()
    idxfti = models.TextField()  # This field type is a guess.
    published = models.BooleanField(default=True)
    searchtext = models.TextField()

    def save(self, *args, **kwargs):
        self.shorttitle = self.title
        self.shorttitle = re.sub("[^a-zA-Z0-9]+", "-", self.shorttitle.lower()).lstrip('-').rstrip('-')
        self.searchtext = strip_tags(markdown.markdown(self.markdown))
        cache.delete('navitems')
        super(Article, self).save(*args, **kwargs)

    def get_absolute_url(self):
        if self.datestamp is not None:
            return '/' + str(self.datestamp.year) + '/' + str(self.datestamp.month) + '/' + str(
                self.datestamp.day) + '/' + self.shorttitle + '/'
        else:
            return ''

    class Meta:
        db_table = u'articles'

    def __unicode__(self):
        return self.title


class Comment(models.Model):
    id = models.AutoField(primary_key=True)
    article = models.ForeignKey(Article)
    datestamp = models.DateTimeField()
    name = models.CharField(max_length=255)
    website = models.CharField(null=True, max_length=255)
    comment = models.TextField()
    ip = models.IPAddressField(null=True)

    def save(self, *args, **kwargs):
        if not self.id:
            self.datestamp = datetime.datetime.now()
        super(Comment, self).save(*args, **kwargs)

    class Meta:
        db_table = u'comments'

    def __unicode__(self):
        return self.name

    def formattedwebsite(self):
        if not self.website.startswith('http'):
            return 'http://' + self.website
        else:
            return self.website


class Location(models.Model):
    timestamp = models.DateTimeField(auto_now_add=True)
    devicetimestamp = models.DateTimeField()
    latitude = models.DecimalField(decimal_places=6, max_digits=9)
    longitude = models.DecimalField(decimal_places=6, max_digits=9)
    accuracy = models.DecimalField(decimal_places=6, max_digits=9)
    geocoding = models.TextField(null=True)

    def save(self, *args, **kwargs):
        if not self.geocoding:
            url = settings.GEOCODE_API_URL.format(self.latitude, self.longitude)
            try:
                r = requests.get(url)
                if r.status_code is 200:
                    self.geocoding = r.text
            except requests.RequestException as e:
                send_mail('Geocode exception on growse.com',
                          'Exception raised while trying to geocode location: {}'.format(e), 'blog@growse.com',
                          'andrew@growse.com')
        super(Location, self).save(*args, **kwargs)

    class Meta:
        db_table = u'locations'
