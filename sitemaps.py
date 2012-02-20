from django.contrib.sitemaps import Sitemap
from blog.models import Article

class BlogSitemap(Sitemap):
    changefreq = "never"
    priority = 0.5

    def items(self):
        return Article.objects.filter(published=True)

    def lastmod(self, obj):
        return obj.datestamp
