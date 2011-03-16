import datetime
from django.contrib.syndication.views import Feed
from blog.models import Article

class RssFeed(Feed):
	title="growse.com"
	link="http://www.growse.com"
	description="ARGLEGARGLEFARGLE"
	def items(self):
		return Article.objects.exclude(datestamp=None).order_by('-datestamp')[:5]

	def item_title(self,item):
		return item.title

	def item_description(self,item):
		return item.body

	def item_link(self,item):
		return 'http://www.growse.com/news/comments/'+item.shorttitle+'/'

	def item_pubdate(self, item):
		 return item.datestamp
