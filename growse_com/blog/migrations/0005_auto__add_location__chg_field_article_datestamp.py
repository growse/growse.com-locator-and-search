# -*- coding: utf-8 -*-
from south.utils import datetime_utils as datetime
from south.db import db
from south.v2 import SchemaMigration
from django.db import models


class Migration(SchemaMigration):

    def forwards(self, orm):
        # Adding model 'Location'
        db.create_table(u'locations', (
            (u'id', self.gf('django.db.models.fields.AutoField')(primary_key=True)),
            ('timestamp', self.gf('django.db.models.fields.DateTimeField')()),
            ('devicetimestamp', self.gf('django.db.models.fields.DateTimeField')()),
            ('latitude', self.gf('django.db.models.fields.DecimalField')(max_digits=9, decimal_places=6)),
            ('longitude', self.gf('django.db.models.fields.DecimalField')(max_digits=9, decimal_places=6)),
            ('accuracy', self.gf('django.db.models.fields.DecimalField')(max_digits=9, decimal_places=6)),
        ))
        db.send_create_signal(u'blog', ['Location'])


        # Changing field 'Article.datestamp'
        db.alter_column(u'articles', 'datestamp', self.gf('django.db.models.fields.DateTimeField')(auto_now_add=True, null=True))

    def backwards(self, orm):
        # Deleting model 'Location'
        db.delete_table(u'locations')


        # Changing field 'Article.datestamp'
        db.alter_column(u'articles', 'datestamp', self.gf('django.db.models.fields.DateTimeField')(null=True))

    models = {
        u'blog.article': {
            'Meta': {'object_name': 'Article', 'db_table': "u'articles'"},
            'datestamp': ('django.db.models.fields.DateTimeField', [], {'auto_now_add': 'True', 'null': 'True', 'blank': 'True'}),
            'description': ('django.db.models.fields.TextField', [], {'null': 'True'}),
            'id': ('django.db.models.fields.AutoField', [], {'primary_key': 'True'}),
            'idxfti': ('django.db.models.fields.TextField', [], {}),
            'markdown': ('django.db.models.fields.TextField', [], {}),
            'published': ('django.db.models.fields.BooleanField', [], {'default': 'True'}),
            'searchtext': ('django.db.models.fields.TextField', [], {}),
            'shorttitle': ('django.db.models.fields.CharField', [], {'unique': 'True', 'max_length': '255'}),
            'title': ('django.db.models.fields.CharField', [], {'max_length': '255'})
        },
        u'blog.comment': {
            'Meta': {'object_name': 'Comment', 'db_table': "u'comments'"},
            'article': ('django.db.models.fields.related.ForeignKey', [], {'to': u"orm['blog.Article']"}),
            'comment': ('django.db.models.fields.TextField', [], {}),
            'datestamp': ('django.db.models.fields.DateTimeField', [], {}),
            'id': ('django.db.models.fields.AutoField', [], {'primary_key': 'True'}),
            'ip': ('django.db.models.fields.IPAddressField', [], {'max_length': '15', 'null': 'True'}),
            'name': ('django.db.models.fields.CharField', [], {'max_length': '255'}),
            'website': ('django.db.models.fields.CharField', [], {'max_length': '255', 'null': 'True'})
        },
        u'blog.location': {
            'Meta': {'object_name': 'Location', 'db_table': "u'locations'"},
            'accuracy': ('django.db.models.fields.DecimalField', [], {'max_digits': '9', 'decimal_places': '6'}),
            'devicetimestamp': ('django.db.models.fields.DateTimeField', [], {}),
            u'id': ('django.db.models.fields.AutoField', [], {'primary_key': 'True'}),
            'latitude': ('django.db.models.fields.DecimalField', [], {'max_digits': '9', 'decimal_places': '6'}),
            'longitude': ('django.db.models.fields.DecimalField', [], {'max_digits': '9', 'decimal_places': '6'}),
            'timestamp': ('django.db.models.fields.DateTimeField', [], {})
        }
    }

    complete_apps = ['blog']