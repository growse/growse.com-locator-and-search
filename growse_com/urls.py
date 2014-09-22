from django.conf.urls import patterns, include
from django.views.generic import TemplateView, RedirectView
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
                       (r'^robots\.txt$', TemplateView.as_view(template_name='robots.txt')),
                       (r'^sitemap\.xml$', 'django.contrib.sitemaps.views.sitemap', {'sitemaps': sitemaps}),
                       (r'^cp/', include(admin.site.urls)),
                       (r'^news/rss/$', RedirectView.as_view(url='/rss/')),
                       (r'^rss/$', RssFeed()),
                       (r'^wxr/$', 'growse_com.blog.views.wxrFeed'),
                       (r'^navlist/(?P<direction>before|since)/(?P<datestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})/$',
                        'growse_com.blog.views.navlist'),
                       (r'^news/comments/(?P<article_shorttitle>.+)/$', 'growse_com.blog.views.article_shorttitle'),
                       (r'^news/comments/(?P<article_shorttitle>.+)/$', 'growse_com.blog.views.article_shorttitle'),
                       (r'^(\d{4})/$', 'growse_com.blog.views.article_bydate'),
                       (r'^(\d{4})/(\d{2})/$', 'growse_com.blog.views.article_bydate'),
                       (r'^(\d{4})/(\d{2})/(\d{2})/$', 'growse_com.blog.views.article_bydate'),
                       (r'^\d{4}/\d{2}/\d{2}/(?P<article_shorttitle>.+)/$', 'growse_com.blog.views.article'),
                       (r'^search/(?P<searchterm>.+)/(?P<page>\d+)/$', 'growse_com.blog.views.search'),
                       (r'^search/(?P<searchterm>.+)/$', 'growse_com.blog.views.search'),
                       (r'^search/$', 'growse_com.blog.views.search'),
                       (r'^locator/$', 'growse_com.blog.views.locator'),
                       (r'^where/$', 'growse_com.blog.views.where'),
                       (r'^where/linestring/$', 'growse_com.blog.views.where_linestring'),
                       (r'^$', 'growse_com.blog.views.article'),
)

urlpatterns += staticfiles_urlpatterns()
