from django.contrib.sitemaps import Sitemap
from blog.models import Article


class BlogSitemap(Sitemap):
    changefreq = "never"
    priority = 0.5

    def items(self):
        return Article.objects.filter(published=True)

    def lastmod(self, obj):
        return obj.datestamp


class FlatSitemap(Sitemap):
    changefreq = "weekly"
    priority = 0.8

    def items(self):
        return '/'

    def location(self, obj):
        return obj


