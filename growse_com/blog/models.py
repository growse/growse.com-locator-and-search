# This is an auto-generated Django model module.
# You'll have to do the following manually to clean this up:
#     * Rearrange models' order
#     * Make sure each model has one field with primary_key=True
# Feel free to rename the models, but don't rename db_table values or field names.
#
# Also note: You'll have to insert the output of 'django-admin.py sqlcustom [appname]'
# into your database.
import math
from decimal import Decimal
from django.core.exceptions import ObjectDoesNotExist
from django.core.mail import send_mail
import json
from django.db import models
from django.utils.html import strip_tags
import datetime
from growse_com import settings
import jsonfield
import re
import markdown
from django.core.cache import cache
import requests
from durationfield.db.models.fields.duration import DurationField


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
    accuracy = models.DecimalField(decimal_places=6, max_digits=12)
    timedelta = DurationField(null=True)
    distance = models.DecimalField(decimal_places=3, max_digits=12, null=True)
    geocoding = jsonfield.JSONField()

    def speed_in_ms(self):
        return self.distance / self.timedelta

    @staticmethod
    def distance_on_unit_sphere(lat1, long1, lat2, long2):

        # Convert latitude and longitude to
        # spherical coordinates in radians.
        degrees_to_radians = math.pi / 180.0

        # phi = 90 - latitude
        phi1 = (90.0 - lat1) * degrees_to_radians
        phi2 = (90.0 - lat2) * degrees_to_radians

        # theta = longitude
        theta1 = long1 * degrees_to_radians
        theta2 = long2 * degrees_to_radians

        # Compute spherical distance from spherical coordinates.

        # For two locations in spherical coordinates
        # (1, theta, phi) and (1, theta, phi)
        # cosine( arc length ) =
        #    sin phi sin phi' cos(theta-theta') + cos phi cos phi'
        # distance = rho * arc length

        cos = (math.sin(phi1) * math.sin(phi2) * math.cos(theta1 - theta2) +
               math.cos(phi1) * math.cos(phi2))
        try:
            cos = max(min(cos, 1.0), -1.0)
        except ValueError as e:
            e.message += " cos=" + str(cos)
            raise e
        arc = math.acos(cos)

        # Remember to multiply arc by the radius of the earth
        # in your favorite set of units to get length.
        return arc

    def save(self, *args, **kwargs):
        try:
            prev = Location.objects.filter(devicetimestamp__lt=self.devicetimestamp).order_by('-devicetimestamp')[
                   :1].get()
            if self.latitude == prev.latitude and self.longitude == prev.longitude:
                self.distance = 0
            else:
                self.distance = 6378100 * Location.distance_on_unit_sphere(float(self.latitude), float(self.longitude),
                                                                           float(prev.latitude),
                                                                           float(prev.longitude))
            self.timedelta = self.devicetimestamp - prev.devicetimestamp
        except ObjectDoesNotExist:
            pass
        # Geocode the things
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

    @staticmethod
    def get_latest():
        last = Location.objects.all().order_by('-timestamp')[:1].get()        
        if 'geonames' in last.geocoding:
            locobj = {'name': last.geocoding['geonames'][0]['name'], 'latitude': last.geocoding['geonames'][0]['lat'],
                      'longitude': last.geocoding['geonames'][0]['lng']}
        return locobj

    class Meta:
        db_table = u'locations'
