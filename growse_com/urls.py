from django.conf.urls import patterns, include
from django.views.generic import TemplateView
from growse_com.blog.rssfeed import RssFeed
from django.contrib import admin
from sitemaps import BlogSitemap, FlatSitemap
from django.contrib.staticfiles.urls import staticfiles_urlpatterns

admin.autodiscover()
sitemaps = {
    'blog': BlogSitemap,
    'flat': FlatSitemap,
}

urlpatterns = patterns('',
                       (r'^robots\.txt$', TemplateView.as_view(template_name='robots.txt')),  # direct_to_template, {'template': 'robots.txt', 'mimetype': 'text/plain'}),
                       (r'^sitemap\.xml$', 'django.contrib.sitemaps.views.sitemap', {'sitemaps': sitemaps}),
                       (r'^cp/', include(admin.site.urls)),
                       (r'^news/rss/$', RssFeed()),
                       (r'^news/comments/(?P<article_shorttitle>.+)/$', 'growse_com.blog.views.article'),
                       (r'^news/$', 'growse_com.blog.views.article'),
                       (r'^photos/$', 'growse_com.blog.views.photos'),
                       (r'^search/(?P<searchterm>.+)/(?P<page>\d+)/$', 'growse_com.blog.views.search'),
                       (r'^search/(?P<searchterm>.+)/$', 'growse_com.blog.views.search'),
                       (r'^search/$', 'growse_com.blog.views.search'),
                       (r'^$', 'growse_com.blog.views.article'),
                       )

urlpatterns += staticfiles_urlpatterns()
