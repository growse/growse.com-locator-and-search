# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models, migrations
import durationfield.db.models.fields.duration
import __builtin__
import jsonfield.fields


class Migration(migrations.Migration):

    dependencies = [
    ]

    operations = [
        migrations.CreateModel(
            name='Article',
            fields=[
                ('id', models.AutoField(serialize=False, primary_key=True)),
                ('datestamp', models.DateTimeField(auto_now_add=True, null=True)),
                ('title', models.CharField(max_length=255)),
                ('shorttitle', models.CharField(unique=True, max_length=255)),
                ('description', models.TextField(null=True)),
                ('markdown', models.TextField()),
                ('idxfti', models.TextField()),
                ('published', models.BooleanField(default=True)),
                ('searchtext', models.TextField()),
            ],
            options={
                'db_table': 'articles',
            },
            bases=(models.Model,),
        ),
        migrations.CreateModel(
            name='Comment',
            fields=[
                ('id', models.AutoField(serialize=False, primary_key=True)),
                ('datestamp', models.DateTimeField()),
                ('name', models.CharField(max_length=255)),
                ('website', models.CharField(max_length=255, null=True)),
                ('comment', models.TextField()),
                ('ip', models.IPAddressField(null=True)),
                ('article', models.ForeignKey(to='blog.Article')),
            ],
            options={
                'db_table': 'comments',
            },
            bases=(models.Model,),
        ),
        migrations.CreateModel(
            name='Location',
            fields=[
                ('id', models.AutoField(verbose_name='ID', serialize=False, auto_created=True, primary_key=True)),
                ('timestamp', models.DateTimeField(auto_now_add=True)),
                ('devicetimestamp', models.DateTimeField()),
                ('latitude', models.DecimalField(max_digits=9, decimal_places=6)),
                ('longitude', models.DecimalField(max_digits=9, decimal_places=6)),
                ('accuracy', models.DecimalField(max_digits=12, decimal_places=6)),
                ('timedelta', durationfield.db.models.fields.duration.DurationField(null=True)),
                ('distance', models.DecimalField(null=True, max_digits=12, decimal_places=3)),
                ('geocoding', jsonfield.fields.JSONField(default=__builtin__.dict)),
            ],
            options={
                'db_table': 'locations',
            },
            bases=(models.Model,),
        ),
    ]
